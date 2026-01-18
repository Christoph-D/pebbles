import type { Plugin } from "@opencode-ai/plugin";
import { tool } from "@opencode-ai/plugin";
import { spawn } from "bun";

/**
 * Pebbles plugin for opencode
 *
 * Put this file into one of these locations:
 *
 * - Project local: .opencode/plugin/pebbles.ts
 * - User global: ~/.opencode/plugin/pebbles.ts
 */

export const PebblesPlugin: Plugin = async ({ $ }) => {
  const prime = await $`peb prime --mcp`.text();

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
          });
          proc.stdin.write(jsonString);
          proc.stdin.end();

          const result = await new Response(proc.stdout).text();
          await proc.exited;
          return result.trim();
        },
      }),
      peb_read: tool({
        description: "Read one or more pebs by ID. Returns full pebs data as JSON.",
        args: {
          id: tool.schema.array(tool.schema.string()).describe("Array of peb IDs to read (e.g., ['peb-xxxx', 'peb-yyyy'])"),
        },
        async execute(args) {
          const proc = spawn(['peb', 'read', ...args.id], {
            stdout: 'pipe',
          });
          const result = await new Response(proc.stdout).text();
          await proc.exited;
          return result.trim();
        },
      }),
      peb_update: tool({
        description: "Update a peb. Optional fields: status (new|in-progress|fixed|wont-fix), title, content, type (bug|feature|epic|task), blocked_by (array of peb IDs)",
        args: {
          id: tool.schema.string().describe("The peb ID to update (e.g., peb-xxxx)"),
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
          });
          const result = await new Response(proc.stdout).text();
          await proc.exited;
          return result.trim();
        },
      }),
      peb_query: tool({
        description: "Query pebs with optional filters (id:peb-xxxx|id:(peb-xxxx|peb-yyyy), status:new|in-progress|fixed|wont-fix|open|closed, type:bug|feature|epic|task, blocked-by:peb-id, --fields:id,title). Returns list of pebs.",
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
          });
          const result = await new Response(proc.stdout).text();
          await proc.exited;
          return result.trim();
        },
      }),
    },
  };
};

export default PebblesPlugin;
