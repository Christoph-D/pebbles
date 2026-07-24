// Version {{.Version}}
//
// IMPORTANT: **This file in .pi/extensions/ is auto-generated**
//
// Changes to this file will be overwritten when you run `peb`.
// If this file is not located in .pi/extensions/, it's safe to modify.
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { getAgentDir } from "@earendil-works/pi-coding-agent";
import { StringEnum } from "@earendil-works/pi-ai";
import { Type } from "typebox";
import { spawn, spawnSync } from "node:child_process";
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

interface FixPebDetails {
	pebId: string;
	workspace: string;
	worktree: string;
	baseRevset: string;
	model?: string;
	changeIds: string[];
	phase: string;
	streamingSummary?: string;
	subagent?: { code: number; turns: number; stopReason?: string };
}

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

/** Walk up from `cwd` to find the nearest `.pi/fix-peb.json`. */
function findProjectFixPebConfig(cwd: string): string | null {
	let dir = cwd;
	while (true) {
		const candidate = path.join(dir, ".pi", "fix-peb.json");
		try {
			if (fs.statSync(candidate).isFile()) return candidate;
		} catch {
			// not present at this level
		}
		const parent = path.dirname(dir);
		if (parent === dir) return null;
		dir = parent;
	}
}

/** Merge user-global then project-local fix-peb config over the defaults. */
function loadFixPebConfig(cwd: string, agentDir: string): FixPebConfig {
	const merged: FixPebConfig = { ...DEFAULT_FIX_PEB_CONFIG };
	const readInto = (file: string) => {
		let raw: string;
		try {
			raw = fs.readFileSync(file, "utf8");
		} catch {
			return;
		}
		let obj: Record<string, unknown>;
		try {
			obj = JSON.parse(raw);
		} catch {
			return; // malformed; ignore
		}
		if (typeof obj["base_revset"] === "string") merged.baseRevset = obj["base_revset"];
		if (typeof obj["worktree_init"] === "string") merged.worktreeInit = obj["worktree_init"];
		if (typeof obj["subagent_model"] === "string") merged.subagentModel = obj["subagent_model"];
		if (typeof obj["commit_message"] === "string") merged.commitMessage = obj["commit_message"];
		if (typeof obj["timeout_ms"] === "number") merged.timeoutMs = obj["timeout_ms"];
		if (typeof obj["max_parallel"] === "number") merged.maxParallel = Math.max(1, Math.floor(obj["max_parallel"]));
	};
	readInto(path.join(agentDir, "fix-peb.json"));
	const projectPath = findProjectFixPebConfig(cwd);
	if (projectPath) readInto(projectPath);
	return merged;
}

/** Simple async semaphore to cap concurrent subagents across fix_peb calls. */
class Semaphore {
	private available: number;
	private readonly waiters: Array<() => void> = [];
	constructor(max: number) {
		this.available = max;
	}
	acquire(): Promise<void> {
		if (this.available > 0) {
			this.available--;
			return Promise.resolve();
		}
		return new Promise<void>((resolve) => this.waiters.push(resolve));
	}
	release(): void {
		const next = this.waiters.shift();
		if (next) next();
		else this.available++;
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

function tryKill(proc: { kill: (sig?: string) => boolean }, sig: string) {
	try {
		proc.kill(sig);
	} catch {
		// ignore
	}
}

/** Spawn a subagent `pi` process in JSON print mode and collect its result. */
function runSubagent(opts: {
	cwd: string;
	model?: string;
	prompt: string;
	timeoutMs: number;
	signal: AbortSignal | undefined;
	onProgress?: (text: string, turns: number) => void;
}): Promise<SubagentResult> {
	return new Promise<SubagentResult>((resolve) => {
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
				if (opts.onProgress && text) opts.onProgress(result.summary, result.turns);
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
			return { content: [{ type: "text", text }] };
		},
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
			return { content: [{ type: "text", text }] };
		},
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
			return { content: [{ type: "text", text }] };
		},
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
			return { content: [{ type: "text", text }] };
		},
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
			return { content: [{ type: "text", text }] };
		},
	});

	// ---- fix_peb: delegate fixing one peb to an isolated subagent in a jj worktree ----
	const fixPebConfig = loadFixPebConfig(process.cwd(), getAgentDir());
	const fixPebSem = new Semaphore(fixPebConfig.maxParallel);
	const jj = (args: string[], opts: { cwd: string }) => pi.exec("jj", args, opts);

	pi.registerTool({
		name: "fix_peb",
		label: "Peb Fix (subagent)",
		promptSnippet: "Delegate fixing one peb to an isolated subagent in a throwaway jj worktree",
		promptGuidelines: [
			"Use fix_peb to hand off fixing a single peb to an isolated subagent that works in its own temporary jj worktree and commits there. You may call fix_peb several times in one turn to fix multiple pebs in parallel. The tool cleans up the worktree itself and only reports success/failure and the resulting commit change ids; it does not merge or push anything.",
		],
		description: [
			"Delegate fixing a single peb to an isolated subagent. The tool reads the peb, creates a temporary jj worktree off the configured base revset (default 'main'), optionally runs a worktree-init script, spawns a subagent (no extensions) that fixes the peb and commits its work with `jj commit`, captures the new commit change ids, then forgets the workspace and deletes the temp dir.",
			`Arguments: peb_id (e.g., ${pebbleIDPattern}), optional extra_prompt appended to the subagent instructions.`,
			"Call fix_peb multiple times in one turn to fix several pebs in parallel. Returns success/failure and the new commit change ids; does not merge or push.",
		].join(" "),
		parameters: Type.Object({
			peb_id: Type.String({ description: `The peb ID to fix (e.g., ${pebbleIDPattern})` }),
			extra_prompt: Type.Optional(
				Type.String({ description: "Optional extra instructions appended to the subagent prompt" }),
			),
		}),
		async execute(_toolCallId, params, signal, onUpdate, ctx) {
			const cfg = fixPebConfig;
			const pebId = String(params.peb_id || "").trim();
			if (!pebId) throw new Error("fix_peb requires a peb_id");

			const details: FixPebDetails = {
				pebId,
				workspace: "",
				worktree: "",
				baseRevset: cfg.baseRevset,
				model: undefined,
				changeIds: [],
				phase: "reading",
			};
			const emit = (text: string, patch?: Partial<FixPebDetails>) => {
				if (patch) Object.assign(details, patch);
				onUpdate?.({ content: [{ type: "text", text }], details });
			};

			await fixPebSem.acquire();
			let tmpDir = "";
			let workspaceName = "";
			let worktreePath = "";
			let workspaceCreated = false;
			try {
				// 1. Read the peb.
				let peb: { id: string; title: string; type?: string; status?: string; content?: string };
				try {
					peb = pebJson(["read", pebId]);
				} catch (e) {
					throw new Error(`Could not read peb ${pebId}: ${(e as Error).message}`);
				}

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
				details.model = model;

				// 2. Temp dir + jj workspace off the base revset.
				tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "peb-fix-"));
				workspaceName = peb.id;
				worktreePath = path.join(tmpDir, peb.id);
				emit(`Creating jj workspace "${workspaceName}" from revset "${cfg.baseRevset}"...`, {
					phase: "worktree",
					workspace: workspaceName,
					worktree: worktreePath,
				});
				const addRes = await jj(["workspace", "add", "-r", cfg.baseRevset, worktreePath], { cwd: ctx.cwd });
				if (addRes.code !== 0) {
					throw new Error(
						`jj workspace add failed (code ${addRes.code}): ${(addRes.stderr || addRes.stdout).trim() || "unknown error"}`,
					);
				}
				workspaceCreated = true;

				// Anchor for reporting new commits: the base commit the worktree sits on.
				const anchorRes = await jj(["log", "-r", "@-", "--no-graph", "-T", "change_id.short()"], { cwd: worktreePath });
				const anchorId = (anchorRes.stdout || "").trim().split(/\s+/)[0] || "";

				// 3. Optional worktree init (runs in main repo cwd, $1 = worktree path).
				if (cfg.worktreeInit) {
					emit(`Running worktree init: ${cfg.worktreeInit}`, { phase: "init" });
					const initRes = await pi.exec("sh", ["-c", cfg.worktreeInit, "sh", worktreePath], { cwd: ctx.cwd });
					if (initRes.code !== 0) {
						throw new Error(
							`worktree init failed (code ${initRes.code}): ${(initRes.stderr || initRes.stdout).trim() || "unknown error"}`,
						);
					}
				}

				// 4. Spawn the subagent.
				const prompt = buildFixPrompt(peb, commitMsg, params.extra_prompt);
				emit(`Running subagent in ${worktreePath}${model ? ` (model ${model})` : ""}...`, { phase: "subagent" });
				const sub = await runSubagent({
					cwd: worktreePath,
					model,
					prompt,
					timeoutMs: cfg.timeoutMs,
					signal,
					onProgress: (text, turns) =>
						emit(`subagent (turn ${turns})...\n${text.slice(0, 600)}`, { phase: "subagent", streamingSummary: text }),
				});

				// 5. Capture new commit change ids (best-effort, before teardown).
				let changeIds: string[] = [];
				if (anchorId) {
					try {
						const logRes = await jj(
							["log", "-r", `${anchorId}..@-`, "--no-graph", "-T", "change_id.short() ++ ' ' ++ description.first_line()"],
							{ cwd: worktreePath },
						);
						changeIds = (logRes.stdout || "")
							.split("\n")
							.map((l) => l.trim())
							.filter(Boolean);
					} catch {
						// best-effort; ignore
					}
				}
				details.changeIds = changeIds;

				const failed = sub.code !== 0 || sub.stopReason === "error" || sub.stopReason === "aborted";
				if (failed) {
					const reason = sub.errorMessage || sub.stderr.trim() || sub.summary || `subagent exited with code ${sub.code}`;
					throw new Error(`Subagent failed to fix ${pebId} (${peb.title}): ${reason}`);
				}

				emit(`fixed ${pebId}: ${changeIds.length} commit(s)`, {
					phase: "done",
					subagent: { code: sub.code, turns: sub.turns, stopReason: sub.stopReason },
				});
				const commitLines = changeIds.length
					? `\n\nNew commits (jj change ids):\n${changeIds.map((c) => `- ${c}`).join("\n")}`
					: "\n\n(no new commits detected — the subagent may not have committed)";
				return {
					content: [
						{
							type: "text",
							text: `Successfully fixed ${peb.id} (${peb.title}) via subagent.${commitLines}\n\nSubagent summary:\n${sub.summary || "(no summary)"}`,
						},
					],
					details,
				};
			} finally {
				// 6. Always tear down the workspace and temp dir (even on error/abort).
				if (workspaceCreated) {
					try {
						await jj(["workspace", "forget", workspaceName], { cwd: ctx.cwd });
					} catch {
						// ignore — best-effort cleanup
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
		},
	});
}
