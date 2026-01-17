import type { Plugin } from "@opencode-ai/plugin";

/**
 * Pebbles prime plugin for opencode
 *
 * Put this file into one of these locations:
 *
 * - Project local: .opencode/plugin/pebbles-prime.ts
 * - User global: ~/.opencode/plugin/pebbles-prime.ts
 */

export const PebblesPrimePlugin: Plugin = async ({ $ }) => {
  const prime = await $`peb prime`.text();

  return {
    "experimental.chat.system.transform": async (_, output) => {
      output.system.push(prime);
    },
    "experimental.session.compacting": async (_, output) => {
      output.context.push(prime);
    },
  };
};

export default PebblesPrimePlugin;
