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

> **Important:** `peb init` and `peb cleanup` are the only commands designed for human use. All other commands (`peb new`, `peb read`, `peb update`, `peb query`, `peb prime`) are designed for AI agents and are not human-friendly. These commands use JSON input/output and are optimized for programmatic use.

### 2. Create a New Task

```bash
echo '{"title":"Fix authentication bug","content":"Users cannot log in","type":"bug"}' | peb new
```

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

To use pebbles with opencode, you need to add the pebbles plugin. This plugin provides:
- **MCP Server** - Tools (`peb_new`, `peb_read`, `peb_update`, `peb_query`) for direct agent integration
- **Agent Instructions** - Automatically injected via `peb prime --mcp`

### Setup

Copy [the plugin file](.opencode/plugin/pebbles.ts) to one of these locations:

- **Project-specific (recommended):** `.opencode/plugin/pebbles.ts`
- **User global (all projects):** `~/.opencode/plugin/pebbles.ts`

### How It Works

The plugin automatically:
1. Runs `peb prime --mcp` to get agent instructions with MCP server tools
2. Injects these instructions into the opencode chat system
3. Provides MCP tools (`peb_new`, `peb_read`, `peb_update`, `peb_query`) for direct access
4. Ensures instructions persist during session compaction
5. Agents automatically track work without manual prompts using the provided tools

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

**Note:** This automatic tracking only works when the pebbles plugin is installed.

### Other Coding Agents

Pebbles can work with any coding agent that supports running shell commands. Show your agent the output of `peb prime` to teach it how to use pebbles.

## Command Reference

### Human Commands

#### `peb init [--opencode]`
Initialize pebbles in current directory (creates `.pebbles/`). With `--opencode` flag, also installs the opencode MCP plugin (creates `.opencode/plugin/pebbles.ts`)

#### `peb cleanup`
Remove all pebbles data (deletes `.pebbles/` directory)

### AI Agent Commands

> **Note:** The following commands are designed for AI agents and use JSON for input/output. They are not intended for human use.

#### `peb new`
Create a new task from JSON via stdin
```bash
echo '{"title":"Fix bug","content":"Description...","type":"bug"}' | peb new
```

#### `peb read <id> [<id> ...]`
Display full task details as JSON (accepts one or more IDs)
```bash
peb read peb-ab12
peb read peb-ab12 peb-cd34 peb-ef56
```

#### `peb update <id> <json>`
Update task fields
```bash
peb update peb-ab12 '{"status":"in-progress"}'
peb update peb-ab12 '{"title":"New title"}'
```

#### `peb query [filters]`
Search and list tasks
```bash
peb query                              # List all
peb query id:peb-ab12                  # Show specific peb
peb query id:(peb-ab12|peb-cd34)       # Show multiple pebs
peb query status:new                   # New tasks only
peb query type:bug                     # Bugs only
peb query status:new type:bug          # New bugs only
peb query --fields id,title status:new # Output specific fields
```

#### `peb prime [--mcp]`
Output agent instructions. With `--mcp` flag, outputs instructions formatted for MCP server integration.

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

## Acknowledgments

- Inspired by [github.com/hmans/beans](https://github.com/hmans/beans) and [github.com/steveyegge/beads/](https://github.com/steveyegge/beads/)
