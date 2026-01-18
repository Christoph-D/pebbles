# Pebbles - Task Tracker for Coding Agents

**Pebbles** (`peb`) is a lightweight agent-first command-line task tracking tool optimized for [opencode](https://opencode.ai) but compatible with any coding agent that can execute shell commands. It helps AI agents track tasks, bugs, features, and epics as they work on your codebase.

## Why Pebbles?

When working with AI coding agents, you'll often have multiple tasks, bugs, and features being worked on simultaneously. Pebbles provides:

- **Task tracking** - Create, read, update, and query tasks as JSON objects
- **Dependency management** - Link related work with `blocked-by` relationships
- **Status tracking** - Track progress through `new` → `in-progress` → `fixed` lifecycle
- **File-based storage** - All tasks stored as individual markdown files in `.pebbles/` directory

## Quick Start

### 1. Initialize Pebbles

```bash
peb init
```

This creates a `.pebbles/` directory with a config file.

### 2. Create a New Task

```bash
echo '{"title":"Fix authentication bug","content":"Users cannot log in","type":"bug"}' | peb new
```

Output: `Created new pebble peb-ab12 in .pebbles/peb-ab12--fix-authentication-bug.md`

### 3. Query Tasks

```bash
# List all tasks
peb query

# Find new bugs
peb query status:new type:bug
```

### 4. Read a Task

```bash
peb read peb-ab12
```

### 5. Update Status

```bash
peb update peb-ab12 '{"status":"in-progress"}'
peb update peb-ab12 '{"status":"fixed"}'
```

## Opencode Integration

To use pebbles with opencode, you need to add the pebbles-prime plugin. This plugin provides task tracking instructions to the opencode agent.

### Setup

Create the plugin file at one of these locations:

**Project-specific (recommended):**
```bash
mkdir -p .opencode/plugin
```

Then create `.opencode/plugin/pebbles-prime.ts`:

```typescript
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
```

**User global (all projects):**
```bash
mkdir -p ~/.opencode/plugin
# Create ~/.opencode/plugin/pebbles-prime.ts with the same content
```

### How It Works

The plugin automatically:
1. Runs `peb prime` to get agent instructions
2. Injects these instructions into the opencode chat system
3. Ensures instructions persist during session compaction
4. Agents automatically track work without manual prompts

## Using Pebbles with Coding Agents

### Opencode Agent Workflow

With the plugin installed, opencode agents use peb automatically to track their work.

### The Agent Workflow

Opencode agents use peb automatically to track their work:

1. **Before starting work** - Agent creates a peb with task details
2. **While working** - Agent updates status to `in-progress`
3. **After completion** - Agent marks peb as `fixed`

### Example Session

```
User: Add a user profile page

Agent: I'll create a task for this and start working on it.
[Creates peb: "Add user profile page" with status "new"]
[Updates peb to "in-progress"]
[Implements the feature]
[Updates peb to "fixed"]

Task peb-xyz1 marked as fixed!
```

**Note:** This automatic tracking only works when the pebbles-prime plugin is installed.

### Creating Dependent Tasks

When tasks depend on each other, use `blocked-by`:

```bash
# Create the blocking task
echo '{"title":"Setup database","content":"Create database schema","type":"task"}' | peb new

# Create the dependent task
echo '{"title":"Build API","content":"Create REST endpoints","type":"task","blocked-by":["peb-ab12"]}' | peb new
```

### Querying Task Status

Check what's pending:

```bash
# Show all open (new or in-progress) tasks
peb query status:open

# Show what's blocked by a specific task
peb query blocked-by:peb-ab12
```

### Other Coding Agents

Pebbles can work with any coding agent that supports running shell commands. Show your agent the output of `peb prime` to teach it how to use pebbles.

## Command Reference

### `peb init`
Initialize pebbles in current directory (creates `.pebbles/`)

### `peb new`
Create a new task from JSON via stdin
```bash
echo '{"title":"Fix bug","content":"Description...","type":"bug"}' | peb new
```

### `peb read <id> [<id> ...]`
Display full task details as JSON (accepts one or more IDs)
```bash
peb read peb-ab12
peb read peb-ab12 peb-cd34 peb-ef56
```

### `peb update <id> <json>`
Update task fields
```bash
peb update peb-ab12 '{"status":"in-progress"}'
peb update peb-ab12 '{"title":"New title"}'
```

### `peb query [filters]`
Search and list tasks
```bash
peb query                              # List all
peb query status:new                   # New tasks only
peb query type:bug                     # Bugs only
peb query status:new type:bug          # New bugs only
peb query --fields id,title status:new # Output specific fields
```

### `peb prime`
Output agent instructions

## Data Model

Each peb has:

- **id**: Unique identifier (e.g., `peb-ab12`)
- **title**: Short description
- **type**: `bug`, `feature`, `epic`, or `task`
- **status**: `new`, `in-progress`, `fixed`, or `wont-fix`
- **blocked-by**: List of peb IDs this task depends on
- **content**: Markdown description
- **created/changed**: Timestamps

## Status Shorthands

- `status:open` - Matches `new` OR `in-progress`
- `status:closed` - Matches `fixed` OR `wont-fix`

## Configuration

Pebbles is configured via `.pebbles/config.toml`:

```toml
prefix = "peb"      # ID prefix
id_length = 4       # Length of random ID portion
```

## Storage

All tasks are stored as individual markdown files in `.pebbles/`:

```markdown
---
id: peb-ab12
title: Fix authentication bug
type: bug
status: new
created: 2026-01-18T12:00:00-08:00
changed: 2026-01-18T12:00:00-08:00
---
Users cannot log in if their name is "null".
```

## Building from Source

```bash
# Clone the repository
git clone <repository>
cd pebbles

# Build
make

# Run
./bin/peb <command>

# Or install to PATH
make install
peb <command>
```

## Testing

```bash
make test              # Run tests
make test-coverage     # Run with coverage report
```

## License

AGPL-3.0. See LICENSE file for details.
