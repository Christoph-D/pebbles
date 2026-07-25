// Version {{.Version}}
//
// IMPORTANT: **This file in .pi/extensions/ is auto-generated**
//
// Changes to this file will be overwritten when you run `peb`.
// If this file is not located in .pi/extensions/, it's safe to modify.
import type { ExtensionAPI, Theme } from "@earendil-works/pi-coding-agent";
import { CONFIG_DIR_NAME, getAgentDir } from "@earendil-works/pi-coding-agent";
import { StringEnum } from "@earendil-works/pi-ai";
import { Text } from "@earendil-works/pi-tui";
import { Type } from "typebox";
import { spawn, spawnSync, type ChildProcess } from "node:child_process";
import * as os from "node:os";
import * as fs from "node:fs";
import * as path from "node:path";

/**
 * Pebbles extension for the pi coding agent.
 *
 * Registers peb tools the LLM can call and injects the pebbles "prime" agent
 * instructions into the system prompt so the agent tracks its work as pebs.
 */

interface RunResult {
	stdout: string;
	stderr: string;
	status: number | null;
}

/** Run `peb` synchronously, optionally piping input to stdin. */
function runPeb(args: string[], stdin?: string): RunResult {
	const result = spawnSync("peb", args, {
		input: stdin,
		encoding: "utf8",
		maxBuffer: 10 * 1024 * 1024,
	});
	return {
		stdout: typeof result.stdout === "string" ? result.stdout : "",
		stderr: typeof result.stderr === "string" ? result.stderr : "",
		status: result.status,
	};
}

/** Run `peb` and return trimmed stdout, appending stderr as "Error:" if set. */
function pebOutput(args: string[], stdin?: string): string {
	const { stdout, stderr } = runPeb(args, stdin);
	const out = stdout.trim();
	const err = stderr.trim();
	return err ? `${out}\nError: ${err}` : out;
}

/** Run a `peb` command that must succeed and return parsed JSON (single object). */
function pebJson(args: string[], stdin?: string): any {
	const res = runPeb(args, stdin);
	if (res.status !== 0) {
		throw new Error(res.stderr.trim() || `peb ${args.join(" ")} failed (exit ${res.status})`);
	}
	let parsed: any;
	try {
		parsed = JSON.parse(res.stdout);
	} catch {
		throw new Error(`peb ${args.join(" ")} returned non-JSON: ${res.stdout.trim().slice(0, 200)}`);
	}
	return Array.isArray(parsed) ? parsed[0] : parsed;
}

/**
 * Return the ids of `blockers` that are still open (status `new` or
 * `in-progress`), via a single `peb query 'id:(...)' 'status:open'` lookup.
 * Returns [] when there are no open blockers (or the list is empty).
 */
function openBlockerIds(blockers: string[]): string[] {
	const ids = blockers.filter((b) => typeof b === "string" && b.trim() !== "");
	if (!ids.length) return [];
	const res = runPeb(["query", `id:(${ids.join("|")})`, "status:open"]);
	const open: string[] = [];
	for (const line of (res.stdout || "").split("\n")) {
		const trimmed = line.trim();
		if (!trimmed) continue;
		try {
			const p = JSON.parse(trimmed);
			if (p && typeof p.id === "string") open.push(p.id);
		} catch {
			// ignore non-JSON lines (e.g. trailing whitespace)
		}
	}
	return open;
}

// ----------------------------------------------------------------------------
// fix_peb: delegate fixing a single peb to an isolated subagent in a jj worktree
// ----------------------------------------------------------------------------

interface FixPebConfig {
	baseRevset: string;
	worktreeInit: string | null;
	subagentModel: string | null;
	commitMessage: string;
	timeoutMs: number;
	maxParallel: number;
}

const DEFAULT_FIX_PEB_CONFIG: FixPebConfig = {
	baseRevset: "main",
	worktreeInit: null,
	subagentModel: null,
	commitMessage: "<message>",
	timeoutMs: 30 * 60 * 1000,
	maxParallel: 6,
};

interface UsageStats {
	input: number;
	output: number;
	cacheRead: number;
	cacheWrite: number;
	cost: number;
}

interface SubagentResult {
	code: number;
	summary: string;
	stderr: string;
	turns: number;
	usage: UsageStats;
	stopReason?: string;
	errorMessage?: string;
	model?: string;
}

/**
 * Read fix_peb configuration from the "pebbles.fixPeb" key of pi's settings.
 *
 * Pi settings live in two JSON files, with the project file overriding the
 * global one (matching pi's own nested-merge semantics):
 *   - global:  <agentDir>/settings.json              (e.g. ~/.pi/agent/settings.json)
 *   - project: <cwd>/<CONFIG_DIR_NAME>/settings.json (e.g. .pi/settings.json)
 *
 * Only the `pebbles.fixPeb` sub-object is consumed, e.g.:
 *
 *   {
 *     "pebbles": {
 *       "fixPeb": {
 *         "baseRevset": "main",
 *         "worktreeInit": "cd \"$1\" && pnpm install",
 *         "subagentModel": "anthropic/claude-sonnet-4-5",
 *         "commitMessage": "fix: {title} ({id})",
 *         "timeoutMs": 1800000,
 *         "maxParallel": 4
 *       }
 *     }
 *   }
 *
 * Missing files and malformed JSON are ignored, so the defaults always apply
 * unless a well-formed value overrides them.
 */
function loadFixPebConfig(cwd: string, agentDir: string): FixPebConfig {
	const merged: FixPebConfig = { ...DEFAULT_FIX_PEB_CONFIG };
	const readInto = (file: string) => {
		let raw: string;
		try {
			raw = fs.readFileSync(file, "utf8");
		} catch {
			return;
		}
		let obj: any;
		try {
			obj = JSON.parse(raw);
		} catch {
			return; // malformed; ignore
		}
		const fp = obj?.pebbles?.fixPeb;
		if (!fp || typeof fp !== "object") return;
		if (typeof fp.baseRevset === "string") merged.baseRevset = fp.baseRevset;
		if (typeof fp.worktreeInit === "string") merged.worktreeInit = fp.worktreeInit;
		if (typeof fp.subagentModel === "string") merged.subagentModel = fp.subagentModel;
		if (typeof fp.commitMessage === "string") merged.commitMessage = fp.commitMessage;
		if (typeof fp.timeoutMs === "number") merged.timeoutMs = fp.timeoutMs;
		if (typeof fp.maxParallel === "number") merged.maxParallel = Math.max(1, Math.floor(fp.maxParallel));
	};
	readInto(path.join(agentDir, "settings.json"));
	readInto(path.join(cwd, CONFIG_DIR_NAME, "settings.json"));
	return merged;
}

/** Simple counter to cap concurrent subagents across fix_peb calls.
 *
 *  Unlike an async semaphore, this does NOT queue: when the limit is hit
 *  `tryAcquire` returns false and fix_peb errors out, so the main model
 *  learns about the limit and can retry later rather than blocking. */
class Semaphore {
	private available: number;
	constructor(private readonly max: number) {
		this.available = max;
	}
	/** Take a slot if one is free. Returns true on success, false at capacity. */
	tryAcquire(): boolean {
		if (this.available > 0) {
			this.available--;
			return true;
		}
		return false;
	}
	release(): void {
		if (this.available < this.max) this.available++;
	}
	/** How many slots are currently held. */
	get running(): number {
		return this.max - this.available;
	}
}

/** Build the prompt handed to the subagent. */
function buildFixPrompt(
	peb: { id: string; title: string; type?: string; status?: string; content?: string },
	commitMsg: string,
	extra?: string,
): string {
	const lines: string[] = [];
	lines.push("You are fixing one tracked issue in an isolated jj worktree.");
	lines.push("Work ONLY inside your current directory (the worktree). Do not touch other workspaces or the main repo.");
	lines.push("");
	lines.push(`ISSUE ${peb.id}: ${peb.title}`);
	const meta = [peb.type ? `type: ${peb.type}` : "", peb.status ? `status: ${peb.status}` : ""].filter(Boolean).join(", ");
	if (meta) lines.push(meta);
	lines.push("");
	lines.push(peb.content || "(no description provided)");
	lines.push("");
	lines.push("Steps:");
	lines.push("1. Investigate the codebase and implement the fix the issue describes. Make minimal, correct changes.");
	lines.push("2. Verify your change (build / tests / lint) if the project provides a way to do so.");
	lines.push("3. Commit ALL of your work with this command: ");
	lines.push(`   jj commit -m "${commitMsg}"`);
	lines.push("   Do NOT mention the issue ID (e.g. peb-xxxx) in the commit message.");
	lines.push("4. Do NOT push, merge, rebase, abandon, or open a pull request. Do NOT create or modify any issues.");
	lines.push("5. When finished, reply with a concise summary of the files you changed and what you did.");
	if (extra && extra.trim()) {
		lines.push("");
		lines.push("Additional instructions from the caller:");
		lines.push(extra.trim());
	}
	return lines.join("\n");
}

function tryKill(proc: { kill: (signal?: NodeJS.Signals | number) => boolean }, sig: NodeJS.Signals) {
	try {
		proc.kill(sig);
	} catch {
		// ignore
	}
}

/** Spawn a subagent `pi` process in JSON print mode. Returns immediately with
 *  the child handle and a promise that resolves when the process exits. The
 *  caller drives completion (capturing commits, tearing down, notifying). */
function spawnSubagent(opts: {
	cwd: string;
	model?: string;
	prompt: string;
	timeoutMs: number;
	signal: AbortSignal | undefined;
	/** Fired after each finalized assistant message with the live result snapshot. */
	onProgress?: (result: SubagentResult) => void;
}): { proc: ChildProcess; result: Promise<SubagentResult> } {
	const args = ["--mode", "json", "-p", "--no-session", "--no-extensions", "--no-context-files", "--approve"];
	if (opts.model) args.push("--model", opts.model);
	args.push(opts.prompt);

	const proc = spawn("pi", args, { cwd: opts.cwd, stdio: ["ignore", "pipe", "pipe"] });
	const result: SubagentResult = {
		code: 0,
		summary: "",
		stderr: "",
		turns: 0,
		usage: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0, cost: 0 },
	};

	const done = new Promise<SubagentResult>((resolve) => {
		let stdoutBuf = "";
		let settled = false;
		const finish = (code: number) => {
			if (settled) return;
			settled = true;
			result.code = code;
			resolve(result);
		};

		const processLine = (line: string) => {
			if (!line) return;
			let event: { type?: string; message?: any };
			try {
				event = JSON.parse(line);
			} catch {
				return;
			}
			if (event.type === "message_end" && event.message && event.message.role === "assistant") {
				const msg = event.message;
				result.turns++;
				let text = "";
				for (const part of msg.content ?? []) {
					if (part && part.type === "text" && typeof part.text === "string") text += part.text;
				}
				if (text) result.summary = text;
				const u = msg.usage;
				if (u) {
					result.usage.input += u.input || 0;
					result.usage.output += u.output || 0;
					result.usage.cacheRead += u.cacheRead || 0;
					result.usage.cacheWrite += u.cacheWrite || 0;
					result.usage.cost += (u.cost && u.cost.total) || 0;
				}
				if (msg.model && !result.model) result.model = msg.model;
				if (msg.stopReason) result.stopReason = msg.stopReason;
				if (msg.errorMessage) result.errorMessage = msg.errorMessage;
				opts.onProgress?.(result);
			}
		};

		proc.stdout.on("data", (data) => {
			stdoutBuf += data.toString();
			const lines = stdoutBuf.split("\n");
			stdoutBuf = lines.pop() || "";
			for (const line of lines) processLine(line);
		});
		proc.stderr.on("data", (data) => {
			result.stderr += data.toString();
		});

		const timer = setTimeout(() => {
			result.errorMessage = result.errorMessage || `subagent timed out after ${Math.round(opts.timeoutMs / 1000)}s`;
			tryKill(proc, "SIGTERM");
			setTimeout(() => tryKill(proc, "SIGKILL"), 5000);
		}, opts.timeoutMs);

		const onAbort = () => {
			result.errorMessage = result.errorMessage || "subagent aborted";
			tryKill(proc, "SIGTERM");
			setTimeout(() => tryKill(proc, "SIGKILL"), 5000);
		};
		if (opts.signal) {
			if (opts.signal.aborted) onAbort();
			else opts.signal.addEventListener("abort", onAbort, { once: true });
		}

		proc.on("error", () => {
			result.errorMessage = result.errorMessage || 'failed to spawn subagent (is "pi" on PATH?)';
			clearTimeout(timer);
			finish(1);
		});
		proc.on("close", (code) => {
			clearTimeout(timer);
			if (stdoutBuf) processLine(stdoutBuf);
			finish(code ?? 0);
		});
	});

	return { proc, result: done };
}

/** A background fix_peb job, kept in the registry for the session. */
interface FixJob {
	pebId: string;
	title: string;
	type?: string;
	workspace: string;
	worktree: string;
	tmpDir: string;
	cwd: string;
	baseRevset: string;
	model?: string;
	startedAt: string;
	status: "running" | "succeeded" | "failed";
	proc?: ChildProcess;
	turns: number;
	summary: string;
	changeIds: string[];
	stopReason?: string;
	errorMessage?: string;
}

/** Registry of background fix_peb jobs, keyed by peb id (one job per peb at a time). */
const fixJobs = new Map<string, FixJob>();

/** Set during session_shutdown so completion handlers skip notifications. */
let sessionShuttingDown = false;

// ----------------------------------------------------------------------------
// Tool result rendering
//
// peb tools return plain text. Without a renderer the TUI's fallback always
// shows the full output and ignores the global ctrl+o (app.tools.expand)
// toggle. renderPebResult honors that toggle like pi's built-in tools:
// collapsed shows a compact one-line summary, expanded shows the full output.
// ----------------------------------------------------------------------------

const PEB_SUMMARY_MAX_CHARS = 100;

/** Parse text as a single JSON value or newline-delimited JSON objects. */
function tryParsePebJson(text: string): unknown[] | null {
	try {
		const v = JSON.parse(text);
		return Array.isArray(v) ? v : [v];
	} catch {
		// not a single JSON value; try newline-delimited JSON below
	}
	const objs: unknown[] = [];
	for (const line of text.split("\n")) {
		const s = line.trim();
		if (!s) continue;
		try {
			objs.push(JSON.parse(s));
		} catch {
			return null;
		}
	}
	return objs.length ? objs : null;
}

/** Build a compact one-line summary of a peb tool's text output. */
function summarizePebOutput(text: string): string {
	const trimmed = text.trim();
	const objs = tryParsePebJson(trimmed);
	if (objs) {
		const ids: string[] = [];
		const titles: string[] = [];
		for (const o of objs) {
			if (o && typeof o === "object") {
				const rec = o as Record<string, unknown>;
				if (typeof rec.id === "string") ids.push(rec.id);
				if (typeof rec.title === "string") titles.push(rec.title);
			}
		}
		if (ids.length === 1) return titles[0] ? `${ids[0]}: ${titles[0]}` : ids[0];
		if (ids.length > 1) return `${ids.length} pebs: ${ids.slice(0, 5).join(", ")}${ids.length > 5 ? " …" : ""}`;
	}
	// Plain text: first line that isn't just punctuation/whitespace.
	const meaningful = trimmed
		.split("\n")
		.map((l) => l.trim())
		.find((l) => l && !/^[{}\[\],\s]*$/.test(l));
	return meaningful ?? trimmed.split("\n")[0] ?? "";
}

/** Collapse to one line and cap length for the collapsed view. */
function truncateForSummary(s: string): string {
	const one = s.replace(/\s+/g, " ").trim();
	return one.length > PEB_SUMMARY_MAX_CHARS ? one.slice(0, PEB_SUMMARY_MAX_CHARS - 1) + " …" : one;
}

/**
 * Render a peb tool result for the TUI, honoring the ctrl+o expand toggle.
 * Collapsed shows a compact summary (+ line count when multi-line); expanded
 * shows the full text output. Mirrors how pi's built-in tools behave.
 */
function renderPebResult(
	result: { content: Array<{ type: string; text?: string }> },
	options: { expanded: boolean; isPartial: boolean },
	theme: Theme,
): Text {
	const text = (result.content ?? [])
		.filter((c) => c.type === "text")
		.map((c) => c.text ?? "")
		.join("\n")
		.trim();

	if (!text) {
		return new Text(theme.fg("dim", options.isPartial ? "running…" : "(no output)"), 0, 0);
	}

	const isError = /^error:/i.test(text);
	const color = isError ? "error" : "toolOutput";

	if (options.expanded) {
		return new Text(theme.fg(color, text), 0, 0);
	}

	const lineCount = text.split("\n").length;
	const summary = truncateForSummary(summarizePebOutput(text));
	const hint = lineCount > 1 ? theme.fg("muted", ` (${lineCount} lines)`) : "";
	return new Text(theme.fg(color, summary) + hint, 0, 0);
}

/**
 * Render the fix_peb completion notification for the TUI, honoring the ctrl+o
 * expand toggle (app.tools.expand) the same way the regular peb tool results
 * do (see renderPebResult). The pi TUI applies that toggle to custom messages
 * too, but only a registered message renderer receives the `expanded` flag —
 * the built-in default always shows the full body and ignores it.
 *
 * Collapsed shows a compact one-line summary (the success/failure header line)
 * with a line-count hint; expanded shows the full notification.
 */
function renderFixPebComplete(
	message: { content: string | Array<{ type: string; text?: string }>; details?: { status?: string } },
	options: { expanded: boolean },
	theme: Theme,
): Text {
	const text = (
		Array.isArray(message.content)
			? message.content.filter((c) => c.type === "text").map((c) => c.text ?? "").join("\n")
			: message.content
	).trim();

	if (!text) {
		return new Text(theme.fg("dim", "(no output)"), 0, 0);
	}

	const status = message.details?.status;
	const color = status === "failed" ? "error" : status === "succeeded" ? "success" : "toolOutput";

	if (options.expanded) {
		return new Text(theme.fg(color, text), 0, 0);
	}

	// Collapsed: the first line is already a good summary
	// ("[fix_peb] Background fix for peb-xxxx (...) SUCCEEDED|FAILED.").
	const lineCount = text.split("\n").length;
	const summary = truncateForSummary(text.split("\n")[0] ?? "");
	const hint = lineCount > 1 ? theme.fg("muted", ` (${lineCount} lines)`) : "";
	return new Text(theme.fg(color, summary) + hint, 0, 0);
}

export default async function (pi: ExtensionAPI) {
	// Load the prime prompt and config once at startup. If `peb` is not on PATH
	// or this isn't a pebbles project, the extension stays inert.
	let prime = "";
	let pebbleIDPattern = "peb-xxxx";
	let pebbleIDPattern2 = "peb-yyyy";
	let pebbleIDPattern3 = "peb-zzzz";

	try {
		const { stdout } = runPeb(["prime", "--mcp"]);
		prime = stdout.trim();
	} catch {
		// peb unavailable; prime stays empty
	}

	try {
		const { stdout } = runPeb(["config"]);
		const config = JSON.parse(stdout);
		const prefix: string = config.prefix;
		const idLength: number = config.id_length;
		pebbleIDPattern = `${prefix}-${"x".repeat(idLength)}`;
		pebbleIDPattern2 = `${prefix}-${"y".repeat(idLength)}`;
		pebbleIDPattern3 = `${prefix}-${"z".repeat(idLength)}`;
	} catch {
		// keep defaults
	}

	// Inject the prime prompt into the system prompt on every agent run. This
	// survives compaction because before_agent_start fires again afterward.
	if (prime) {
		pi.on("before_agent_start", async (event) => {
			return {
				systemPrompt: event.systemPrompt ? `${event.systemPrompt}\n\n${prime}` : prime,
			};
		});
	}

	// Render the fix_peb completion notification so it honors the ctrl+o expand
	// toggle, just like the regular peb tool results. Without a registered
	// renderer the built-in default ignores the toggle and always shows the
	// full body.
	pi.registerMessageRenderer("fix-peb-complete", renderFixPebComplete);

	pi.registerTool({
		name: "peb_new",
		label: "Peb New",
		description:
			"Create a new peb (task/bug/feature/epic). Required: title, content. Optional: type (bug|feature|epic|task, default: bug), blocked_by (array of peb IDs)",
		parameters: Type.Object({
			title: Type.String({ description: "Short description of the peb" }),
			content: Type.String({ description: "Markdown description of the peb" }),
			type: Type.Optional(
				StringEnum(["bug", "feature", "epic", "task"], {
					description: "Type: bug, feature, epic, or task (default: bug)",
				}),
			),
			blocked_by: Type.Optional(
				Type.Array(Type.String(), {
					description: "Array of peb IDs that block this peb",
				}),
			),
		}),
		async execute(_toolCallId, params) {
			const json: Record<string, unknown> = {
				title: params.title,
				content: params.content,
			};
			if (params.type) json.type = params.type;
			if (params.blocked_by) json["blocked-by"] = params.blocked_by;
			const text = pebOutput(["new"], JSON.stringify(json));
			return { content: [{ type: "text", text }], details: undefined };
		},
		renderResult: renderPebResult,
	});

	pi.registerTool({
		name: "peb_read",
		label: "Peb Read",
		description: "Read one or more pebs by ID. Returns full pebs data as JSON.",
		parameters: Type.Object({
			id: Type.Array(Type.String(), {
				description: `Array of peb IDs to read (e.g., ['${pebbleIDPattern}', '${pebbleIDPattern2}'])`,
			}),
		}),
		async execute(_toolCallId, params) {
			const text = pebOutput(["read", ...params.id]);
			return { content: [{ type: "text", text }], details: undefined };
		},
		renderResult: renderPebResult,
	});

	pi.registerTool({
		name: "peb_update",
		label: "Peb Update",
		description:
			"Update a peb. Optional fields: status (new|in-progress|fixed|wont-fix), title, content, type (bug|feature|epic|task), blocked_by (array of peb IDs)",
		parameters: Type.Object({
			id: Type.String({ description: `The peb ID to update (e.g., ${pebbleIDPattern})` }),
			status: Type.Optional(
				StringEnum(["new", "in-progress", "fixed", "wont-fix"], {
					description: "Status: new, in-progress, fixed, or wont-fix",
				}),
			),
			title: Type.Optional(Type.String({ description: "Short description of the peb" })),
			content: Type.Optional(Type.String({ description: "Markdown description of the peb" })),
			type: Type.Optional(
				StringEnum(["bug", "feature", "epic", "task"], {
					description: "Type: bug, feature, epic, or task",
				}),
			),
			blocked_by: Type.Optional(
				Type.Array(Type.String(), {
					description: "Array of peb IDs that block this peb",
				}),
			),
		}),
		async execute(_toolCallId, params) {
			const json: Record<string, unknown> = {};
			if (params.status) json.status = params.status;
			if (params.title) json.title = params.title;
			if (params.content) json.content = params.content;
			if (params.type) json.type = params.type;
			if (params.blocked_by) json["blocked-by"] = params.blocked_by;
			const text = pebOutput(["update", params.id, JSON.stringify(json)]);
			return { content: [{ type: "text", text }], details: undefined };
		},
		renderResult: renderPebResult,
	});

	pi.registerTool({
		name: "peb_query",
		label: "Peb Query",
		description: `Query pebs with optional filters (id:${pebbleIDPattern}|id:(${pebbleIDPattern}|${pebbleIDPattern2}), status:new|in-progress|fixed|wont-fix|open|closed, type:bug|feature|epic|task, blocked-by:${pebbleIDPattern}, --fields:id,title). Returns list of pebs.`,
		parameters: Type.Object({
			filters: Type.Optional(
				Type.Array(Type.String(), {
					description: "Array of filters (e.g., ['status:new', 'type:bug'])",
				}),
			),
			fields: Type.Optional(
				Type.Array(Type.String(), {
					description: "Array of fields to output (e.g., ['id', 'title'])",
				}),
			),
		}),
		async execute(_toolCallId, params) {
			const args: string[] = ["query"];
			if (params.fields) args.push("--fields", params.fields.join(","));
			if (params.filters) args.push(...params.filters);
			const text = pebOutput(args);
			return { content: [{ type: "text", text }], details: undefined };
		},
		renderResult: renderPebResult,
	});

	pi.registerTool({
		name: "peb_delete",
		label: "Peb Delete",
		description: "Delete pebs by ID.",
		parameters: Type.Object({
			id: Type.Array(Type.String(), {
				description: `Array of peb IDs to delete (e.g., ['${pebbleIDPattern}', '${pebbleIDPattern2}'])`,
			}),
		}),
		async execute(_toolCallId, params) {
			const text = pebOutput(["delete", ...params.id]);
			return { content: [{ type: "text", text }], details: undefined };
		},
		renderResult: renderPebResult,
	});

	// ---- fix_peb: delegate fixing one peb to an isolated subagent in a jj worktree ----
	const fixPebConfig = loadFixPebConfig(process.cwd(), getAgentDir());
	const fixPebSem = new Semaphore(fixPebConfig.maxParallel);
	const jj = (args: string[], opts: { cwd: string }) => pi.exec("jj", args, opts);

	/** Verify cwd is inside a jj repository; throw a helpful error otherwise.
	 *  fix_peb relies on jj worktrees, so it only works in jj repositories. */
	const requireJjRepo = async (cwd: string): Promise<void> => {
		let res: Awaited<ReturnType<typeof jj>>;
		try {
			res = await jj(["root"], { cwd });
		} catch (e) {
			throw new Error(
				`fix_peb only works in jj repositories, but jj could not be run (${(e as Error).message}). ` +
					`Install jj — see https://docs.jj-vcs.dev/.`,
			);
		}
		if (res.code !== 0) {
			const detail = (res.stderr || res.stdout || "").trim();
			throw new Error(
				`fix_peb only works in jj repositories, but this does not appear to be one` +
					(detail ? `: ${detail}` : "") +
				`. See https://docs.jj-vcs.dev/ to get started.`,
			);
		}
	};

	/** Build the one-shot completion notification and wake the main agent.
	 *  Called exactly once per job, only on finish (success or failure). */
	const notifyFixComplete = (job: FixJob, sub: SubagentResult) => {
		const success = job.status === "succeeded";
		const lines: string[] = [];
		lines.push(
			success
				? `[fix_peb] Background fix for ${job.pebId} (${job.title}) SUCCEEDED.`
				: `[fix_peb] Background fix for ${job.pebId} (${job.title}) FAILED.`,
		);
		if (success && job.changeIds.length) {
			lines.push("New commits (jj change ids):");
			for (const c of job.changeIds) lines.push(`- ${c}`);
			const firstChangeId = job.changeIds[0].split(/\s+/)[0];
			lines.push(
				"Bring these changes into your working copy by rebasing them before the current commit:",
			);
			lines.push(`  jj rebase --source ${firstChangeId} --insert-before @`);
			lines.push(
				"(Only the first change id is needed — --source rebases that change and all of its descendants.)",
			);
		} else if (success) {
			lines.push("(no new commits detected — the subagent may not have committed)");
		}
		if (success) {
			if (job.summary) lines.push("", "Subagent summary:", job.summary);
			lines.push("", `Subagent turns: ${job.turns}`);
		}
		if (!success) {
			const reason = job.errorMessage || sub.stderr.trim() || sub.summary || `subagent exited with code ${sub.code}`;
			lines.push("", `Reason: ${reason}`);
		}
		try {
			pi.sendMessage(
				{
					customType: "fix-peb-complete",
					content: lines.join("\n"),
					display: true,
					details: { pebId: job.pebId, status: job.status, changeIds: job.changeIds },
				},
				{ triggerTurn: true, deliverAs: "followUp" },
			);
		} catch {
			// session may be gone; best-effort
		}
	};

	/** Forget a job's workspace and remove its temp dir. Best-effort. */
	const teardownJob = async (job: FixJob) => {
		if (job.workspace) {
			try {
				await jj(["workspace", "forget", job.workspace], { cwd: job.cwd });
			} catch {
				// ignore — best-effort
			}
		}
		if (job.tmpDir) {
			try {
				fs.rmSync(job.tmpDir, { recursive: true, force: true });
			} catch {
				// ignore
			}
		}
	};

	pi.registerTool({
		name: "fix_peb",
		label: "Peb Fix (background subagent)",
		promptSnippet: "Delegate fixing one peb to a background subagent in a throwaway jj worktree",
		// `description` travels via the tool's function schema (read at call-decision
		// time); `promptGuidelines` are standing bullets in the system prompt. Keep
		// them disjoint: mechanics + hard runtime constraints here, workflow there.
		promptGuidelines: [
			"fix_peb runs as a BACKGROUND job: it returns immediately and notifies you when the subagent finishes (success or failure). Launch several in one turn to fix pebs in parallel — use fix_peb_list to monitor jobs and fix_peb_kill to abort one. fix_peb does not merge or push anything.",
			"After a successful fix, rebase the new commits before your working copy with `jj rebase --source <first-change-id> --insert-before @` (only the first change id is needed — --source rebases it and all descendants) to pull the subagent's work into the main repo.",
		],
		description: [
			"Delegate fixing a single peb to an isolated BACKGROUND subagent. The tool reads the peb, creates a temporary jj worktree off the configured base revset (default 'main'), optionally runs a worktree-init script, spawns a subagent that fixes the peb and commits with `jj commit`, and returns IMMEDIATELY with a job id — it does NOT wait for the subagent. When the subagent finishes (success or failure) the main agent is notified via a message with the new commit change ids, then the worktree is forgotten and removed (commits stay reachable in jj).",
			`Arguments: peb_id (e.g., ${pebbleIDPattern}), optional extra_prompt appended to the subagent instructions.`,
			"CANNOT be used on a peb that has open blockers: every peb in its `blocked-by` list must be fixed (or closed) first.",
		].join(" "),
		parameters: Type.Object({
			peb_id: Type.String({ description: `The peb ID to fix (e.g., ${pebbleIDPattern})` }),
			extra_prompt: Type.Optional(
				Type.String({ description: "Optional extra instructions appended to the subagent prompt" }),
			),
		}),
		async execute(_toolCallId, params, _signal, onUpdate, ctx) {
			const cfg = fixPebConfig;
			const pebId = String(params.peb_id || "").trim();
			if (!pebId) throw new Error("fix_peb requires a peb_id");

			const existing = fixJobs.get(pebId);
			if (existing && existing.status === "running") {
				throw new Error(
					`A fix is already running for ${pebId}. Use fix_peb_list to monitor it or fix_peb_kill to stop it before starting another.`,
				);
			}

			// fix_peb works by creating jj worktrees, so it only works inside jj
			// repositories. Fail fast with an actionable error before doing anything.
			await requireJjRepo(ctx.cwd);

			// 1. Read the peb and reject it if it still has open blockers. Do this
			//    before acquiring a slot or creating a worktree, so a rejected call
			//    is cheap and doesn't hold the concurrency semaphore.
			let peb: {
				id: string;
				title: string;
				type?: string;
				status?: string;
				content?: string;
				"blocked-by"?: unknown[];
			};
			try {
				peb = pebJson(["read", pebId]);
			} catch (e) {
				throw new Error(`Could not read peb ${pebId}: ${(e as Error).message}`);
			}
			const blockers = ((peb["blocked-by"] as unknown[] | undefined) ?? []).filter(
				(b): b is string => typeof b === "string" && b.trim() !== "",
			);
			const openBlockers = openBlockerIds(blockers);
			if (openBlockers.length) {
				throw new Error(
					`${pebId} (${peb.title || ""}) is blocked by ${openBlockers.length} open peb(s): ` +
						`${openBlockers.join(", ")}. ` +
						`fix_peb cannot be used on a blocked peb; fix or close its blockers first, then retry.`,
				);
			}

			const emit = (text: string) => {
				onUpdate?.({ content: [{ type: "text", text }], details: undefined });
			};

			if (!fixPebSem.tryAcquire()) {
				throw new Error(
					`fix_peb concurrency limit reached: at most ${fixPebConfig.maxParallel} background fix_peb job(s) ` +
						`may run at once and all ${fixPebConfig.maxParallel} slot(s) are currently in use ` +
						`(${fixPebSem.running} running). ` +
						`Wait for a running job to complete — you will receive a notification when one finishes, then retry.`,
				);
			}
			let tmpDir = "";
			let workspaceName = "";
			let worktreePath = "";
			let workspaceCreated = false;
			let anchorId = "";
			let spawned = false;
			try {
				// Commit message: substitute {id}/{title} and sanitize to a single, shell-safe line.
				const title = (peb.title || "").replace(/\s+/g, " ").trim();
				const commitMsg = cfg.commitMessage
					.replace(/\{id\}/g, peb.id)
					.replace(/\{title\}/g, title)
					.replace(/\s+/g, " ")
					.replace(/["`$\\]/g, "")
					.trim();

				// Resolve model: explicit config, else the main agent's current model.
				const model = cfg.subagentModel ?? (ctx.model ? `${ctx.model.provider}/${ctx.model.id}` : undefined);

				// 2. Temp dir + jj workspace off the base revset.
				tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "peb-fix-"));
				workspaceName = peb.id;
				worktreePath = path.join(tmpDir, peb.id);
				emit(`Creating jj workspace "${workspaceName}" from revset "${cfg.baseRevset}"...`);
				const addRes = await jj(["workspace", "add", "-r", cfg.baseRevset, worktreePath], { cwd: ctx.cwd });
				if (addRes.code !== 0) {
					throw new Error(
						`jj workspace add failed (code ${addRes.code}): ${(addRes.stderr || addRes.stdout).trim() || "unknown error"}`,
					);
				}
				workspaceCreated = true;

				// Anchor for reporting new commits: the base commit the worktree sits on.
				const anchorRes = await jj(["log", "-r", "@-", "--no-graph", "-T", "change_id.short()"], { cwd: worktreePath });
				anchorId = (anchorRes.stdout || "").trim().split(/\s+/)[0] || "";

				// 3. Optional worktree init (runs in main repo cwd, $1 = worktree path).
				if (cfg.worktreeInit) {
					emit(`Running worktree init: ${cfg.worktreeInit}`);
					const initRes = await pi.exec("sh", ["-c", cfg.worktreeInit, "sh", worktreePath], { cwd: ctx.cwd });
					if (initRes.code !== 0) {
						throw new Error(
							`worktree init failed (code ${initRes.code}): ${(initRes.stderr || initRes.stdout).trim() || "unknown error"}`,
						);
					}
				}

				// 4. Spawn the subagent (non-blocking) and register the job.
				const prompt = buildFixPrompt(peb, commitMsg, params.extra_prompt);
				emit(`Started subagent in ${worktreePath}${model ? ` (model ${model})` : ""}`);

				// Build the job shell first so the stream progress callback can mirror
				// live counters (turns, summary, ...) onto it as they arrive.
				const job: FixJob = {
					pebId,
					title: peb.title,
					type: peb.type,
					workspace: workspaceName,
					worktree: worktreePath,
					tmpDir,
					cwd: ctx.cwd,
					baseRevset: cfg.baseRevset,
					model,
					startedAt: new Date().toISOString(),
					status: "running",
					proc: undefined,
					turns: 0,
					summary: "",
					changeIds: [],
					stopReason: undefined,
					errorMessage: undefined,
				};
				const { proc, result } = spawnSubagent({
					cwd: worktreePath,
					model,
					prompt,
					timeoutMs: cfg.timeoutMs,
					signal: undefined,
					onProgress: (r) => {
						// Mirror live progress so fix_peb_list reflects in-flight jobs.
						job.turns = r.turns;
						if (r.summary) job.summary = r.summary;
						if (r.model && !job.model) job.model = r.model;
						if (r.stopReason) job.stopReason = r.stopReason;
						if (r.errorMessage) job.errorMessage = job.errorMessage || r.errorMessage;
					},
				});
				job.proc = proc;
				spawned = true;
				fixJobs.set(pebId, job);

				// 5. Background completion: capture commits, tear down, notify once.
				void result.then(async (sub) => {
					try {
						job.summary = sub.summary;
						job.turns = sub.turns;
						job.stopReason = sub.stopReason;
						job.errorMessage = job.errorMessage || sub.errorMessage;
						job.proc = undefined;

						// Capture new commit change ids before forgetting the workspace.
						if (anchorId) {
							try {
								const logRes = await jj(
									["log", "-r", `${anchorId}..@-`, "--no-graph", "-T", "change_id.short() ++ ' ' ++ description.first_line()"],
									{ cwd: worktreePath },
								);
								job.changeIds = (logRes.stdout || "")
									.split("\n")
									.map((l) => l.trim())
									.filter(Boolean);
							} catch {
								// best-effort
							}
						}

						const failed = sub.code !== 0 || sub.stopReason === "error" || sub.stopReason === "aborted";
						job.status = failed ? "failed" : "succeeded";

						await teardownJob(job);
					} finally {
						fixPebSem.release();
						if (!sessionShuttingDown) {
							try {
								notifyFixComplete(job, sub);
							} catch {
								// ignore — notification is best-effort
							}
						}
					}
				});

				return {
					content: [
						{
							type: "text",
							text: `Started background fix for ${peb.id} (${peb.title}) in worktree ${worktreePath}${model ? ` (model ${model})` : ""}. You will be notified when the subagent finishes (success or failure). Use fix_peb_list to monitor or fix_peb_kill to abort.`,
						},
					],
					details: {
						pebId,
						workspace: workspaceName,
						worktree: worktreePath,
						baseRevset: cfg.baseRevset,
						model,
						phase: "running",
					},
				};
			} catch (e) {
				// Setup failed before the subagent was spawned: clean up and release
				// the slot. (On a spawned run the completion handler releases.)
				if (!spawned) {
					if (workspaceCreated) {
						try {
							await jj(["workspace", "forget", workspaceName], { cwd: ctx.cwd });
						} catch {
							// ignore
						}
					}
					if (tmpDir) {
						try {
							fs.rmSync(tmpDir, { recursive: true, force: true });
						} catch {
							// ignore
						}
					}
					fixPebSem.release();
				}
				throw e;
			}
		},
		renderResult: renderPebResult,
	});

	pi.registerTool({
		name: "fix_peb_list",
		label: "Peb Fix List",
		promptSnippet: "List running and finished background fix_peb jobs",
		description:
			"List background fix_peb jobs (running and finished) with their status, worktree, subagent summary, and resulting commit change ids. Does not block.",
		parameters: Type.Object({}),
		async execute() {
			const jobs = Array.from(fixJobs.values()).map((j) => ({
				pebId: j.pebId,
				title: j.title,
				type: j.type,
				status: j.status,
				workspace: j.workspace,
				worktree: j.worktree,
				baseRevset: j.baseRevset,
				model: j.model,
				startedAt: j.startedAt,
				turns: j.turns,
				summary: j.summary,
				changeIds: j.changeIds,
				stopReason: j.stopReason,
				errorMessage: j.errorMessage,
			}));
			return {
				content: [
					{
						type: "text",
						text: jobs.length ? JSON.stringify(jobs, null, 2) : "No fix_peb jobs. Use fix_peb to start one.",
					},
				],
				details: undefined,
			};
		},
		renderResult: renderPebResult,
	});

	pi.registerTool({
		name: "fix_peb_kill",
		label: "Peb Fix Kill",
		promptSnippet: "Kill a running background fix_peb subagent",
		description:
			"Abort a running background fix_peb subagent by peb id. Sends SIGTERM to the subagent; its workspace is torn down and a completion (failure) notification is delivered when it exits. Throws if the job is not running.",
		parameters: Type.Object({
			peb_id: Type.String({ description: `The peb ID whose fix subagent to kill (e.g., ${pebbleIDPattern})` }),
		}),
		async execute(_toolCallId, params) {
			const pebId = String(params.peb_id || "").trim();
			const job = fixJobs.get(pebId);
			if (!job) throw new Error(`No fix_peb job found for ${pebId}. Use fix_peb_list to see jobs.`);
			if (job.status !== "running" || !job.proc) {
				throw new Error(`Fix for ${pebId} is not running (status: ${job.status}).`);
			}
			job.errorMessage = job.errorMessage || "killed by main agent via fix_peb_kill";
			const proc = job.proc;
			tryKill(proc, "SIGTERM");
			setTimeout(() => {
				if (job.proc) tryKill(job.proc, "SIGKILL");
			}, 5000);
			return {
				content: [
					{
						type: "text",
						text: `Sent SIGTERM to the subagent for ${pebId}. It will be torn down and a failure notification will follow when it exits.`,
					},
				],
				details: undefined,
			};
		},
		renderResult: renderPebResult,
	});

	// Kill any still-running subagents and tear down their workspaces on exit.
	// Children do not need to survive reload/restart, so no detached/unref.
	pi.on("session_shutdown", async () => {
		if (sessionShuttingDown) return;
		sessionShuttingDown = true;
		for (const job of fixJobs.values()) {
			if (job.status === "running" && job.proc) tryKill(job.proc, "SIGTERM");
		}
		await Promise.all(Array.from(fixJobs.values()).map((job) => teardownJob(job)));
	});
}
