// Version {{.Version}}
//
// IMPORTANT: **This file in .opencode/plugin/ is auto-generated**
//
// Changes to this file will be overwritten when you run `peb`.
// If this file is not located in .opencode/plugin/, it's safe to modify.
import type { Plugin } from "@opencode-ai/plugin";
import { tool } from "@opencode-ai/plugin";
import { spawn } from "bun";

/**
 * Pebbles plugin for opencode
 */

export const PebblesPlugin: Plugin = async ({ $ }) => {
  const prime = await $`peb prime --mcp`.text();

  const configProc = spawn(['peb', 'config'], {
    stdout: 'pipe',
    stderr: 'pipe',
  });

  const [configText, configError] = await Promise.all([
    new Response(configProc.stdout).text(),
    new Response(configProc.stderr).text(),
  ]);
  await configProc.exited;

  if (configError.trim()) {
    throw new Error(`Failed to get config: ${configError.trim()}`);
  }

  const config = JSON.parse(configText);
  const prefix = config.prefix;
  const idLength = config.id_length;
  const pebbleIDPattern = `${prefix}-${"x".repeat(idLength)}`;
  const pebbleIDPattern2 = `${prefix}-${"y".repeat(idLength)}`;
  const pebbleIDPattern3 = `${prefix}-${"z".repeat(idLength)}`;

  return {
    "experimental.chat.system.transform": async (_, output) => {
      output.system.push(prime);
    },
    "experimental.session.compacting": async (_, output) => {
      output.context.push(prime);
    },
    tool: {
      peb_new: tool({
        description: "Create a new peb (task/bug/feature/epic). Required: title, content. Optional: type (bug|feature|epic|task, default: bug), blocked-by (array of peb IDs)",
        args: {
          title: tool.schema.string().describe("Short description of the peb"),
          content: tool.schema.string().describe("Markdown description of the peb"),
          type: tool.schema.string().optional().describe("Type: bug, feature, epic, or task (default: bug)"),
          blocked_by: tool.schema.array(tool.schema.string()).optional().describe("Array of peb IDs that block this peb"),
        },
        async execute(args) {
          const json: Record<string, unknown> = { title: args.title, content: args.content };
          if (args.type) json.type = args.type;
          if (args.blocked_by) json["blocked-by"] = args.blocked_by;

          const jsonString = JSON.stringify(json);
          const proc = spawn(['peb', 'new'], {
            stdin: 'pipe',
            stdout: 'pipe',
            stderr: 'pipe',
          });
          proc.stdin.write(jsonString);
          proc.stdin.end();

          const [stdout, stderr] = await Promise.all([
            new Response(proc.stdout).text(),
            new Response(proc.stderr).text(),
          ]);
          await proc.exited;
          const result = stdout.trim();
          const errorOutput = stderr.trim();
          return errorOutput ? `${result}\n$Error: {errorOutput}` : result;
        },
      }),
      peb_read: tool({
        description: "Read one or more pebs by ID. Returns full pebs data as JSON.",
        args: {
          id: tool.schema.array(tool.schema.string()).describe(`Array of peb IDs to read (e.g., ['${pebbleIDPattern}', '${pebbleIDPattern2}'])`),
        },
        async execute(args) {
          const proc = spawn(['peb', 'read', ...args.id], {
            stdout: 'pipe',
            stderr: 'pipe',
          });
          const [stdout, stderr] = await Promise.all([
            new Response(proc.stdout).text(),
            new Response(proc.stderr).text(),
          ]);
          await proc.exited;
          const result = stdout.trim();
          const errorOutput = stderr.trim();
          return errorOutput ? `${result}\n$Error: {errorOutput}` : result;
        },
      }),
      peb_update: tool({
        description: "Update a peb. Optional fields: status (new|in-progress|fixed|wont-fix), title, content, type (bug|feature|epic|task), blocked_by (array of peb IDs)",
        args: {
          id: tool.schema.string().describe(`The peb ID to update (e.g., ${pebbleIDPattern})`),
          status: tool.schema.string().optional().describe("Status: new, in-progress, fixed, or wont-fix"),
          title: tool.schema.string().optional().describe("Short description of the peb"),
          content: tool.schema.string().optional().describe("Markdown description of the peb"),
          type: tool.schema.string().optional().describe("Type: bug, feature, epic, or task"),
          blocked_by: tool.schema.array(tool.schema.string()).optional().describe("Array of peb IDs that block this peb"),
        },
        async execute(args) {
          const json: Record<string, unknown> = {};
          if (args.status) json.status = args.status;
          if (args.title) json.title = args.title;
          if (args.content) json.content = args.content;
          if (args.type) json.type = args.type;
          if (args.blocked_by) json["blocked-by"] = args.blocked_by;

          const jsonString = JSON.stringify(json);
          const proc = spawn(['peb', 'update', args.id, jsonString], {
            stdout: 'pipe',
            stderr: 'pipe',
          });
          const [stdout, stderr] = await Promise.all([
            new Response(proc.stdout).text(),
            new Response(proc.stderr).text(),
          ]);
          await proc.exited;
          const result = stdout.trim();
          const errorOutput = stderr.trim();
          return errorOutput ? `${result}\n$Error: {errorOutput}` : result;
        },
      }),
      peb_query: tool({
        description: `Query pebs with optional filters (id:${pebbleIDPattern}|id:(${pebbleIDPattern}|${pebbleIDPattern2}), status:new|in-progress|fixed|wont-fix|open|closed, type:bug|feature|epic|task, blocked-by:${pebbleIDPattern}, --fields:id,title). Returns list of pebs.`,
        args: {
          filters: tool.schema.array(tool.schema.string()).optional().describe("Array of filters (e.g., ['status:new', 'type:bug'])"),
          fields: tool.schema.array(tool.schema.string()).optional().describe("Array of fields to output (e.g., ['id', 'title'])"),
        },
        async execute(args) {
          const cmdArgs: string[] = ['peb', 'query'];
          if (args.fields) {
            cmdArgs.push('--fields', args.fields.join(','));
          }
          if (args.filters) {
            cmdArgs.push(...args.filters);
          }

          const proc = spawn(cmdArgs, {
            stdout: 'pipe',
            stderr: 'pipe',
          });
          const [stdout, stderr] = await Promise.all([
            new Response(proc.stdout).text(),
            new Response(proc.stderr).text(),
          ]);
          await proc.exited;
          const result = stdout.trim();
          const errorOutput = stderr.trim();
          return errorOutput ? `${result}\n$Error: {errorOutput}` : result;
        },
      }),
      peb_delete: tool({
        description: "Delete pebs by ID.",
        args: {
          id: tool.schema.array(tool.schema.string()).describe(`Array of peb IDs to delete (e.g., ['${pebbleIDPattern}', '${pebbleIDPattern2}'])`),
        },
        async execute(args) {
          const proc = spawn(['peb', 'delete', ...args.id], {
            stdout: 'pipe',
            stderr: 'pipe',
          });
          const [stdout, stderr] = await Promise.all([
            new Response(proc.stdout).text(),
            new Response(proc.stderr).text(),
          ]);
          await proc.exited;
          const result = stdout.trim();
          const errorOutput = stderr.trim();
          return errorOutput ? `${result}\n$Error: {errorOutput}` : result;
        },
      }),
    },
  };
};

export default PebblesPlugin;
