// Version {{.Version}}
//
// IMPORTANT: **This file in .pi/extensions/ is auto-generated**
//
// Changes to this file will be overwritten when you run `peb`.
// If this file is not located in .pi/extensions/, it's safe to modify.
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { StringEnum } from "@earendil-works/pi-ai";
import { Type } from "typebox";
import { spawnSync } from "node:child_process";

/**
 * Pebbles extension for the pi coding agent.
 *
 * Registers peb tools the LLM can call and injects the pebbles "prime" agent
 * instructions into the system prompt so the agent tracks its work as pebs.
 */

interface RunResult {
	stdout: string;
	stderr: string;
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
	};
}

/** Run `peb` and return trimmed stdout, appending stderr as "Error:" if set. */
function pebOutput(args: string[], stdin?: string): string {
	const { stdout, stderr } = runPeb(args, stdin);
	const out = stdout.trim();
	const err = stderr.trim();
	return err ? `${out}\nError: ${err}` : out;
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
}
