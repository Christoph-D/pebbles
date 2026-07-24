# Pebbles - Task Tracker for Coding Agents

[![CI](https://github.com/Christoph-D/pebbles/actions/workflows/test.yml/badge.svg)](https://github.com/Christoph-D/pebbles/actions/workflows/test.yml)

**Pebbles** (`peb`) is a lightweight agent-first command-line task tracking tool
optimized for [opencode](https://opencode.ai) but compatible with any coding
agent that can execute shell commands.

It's a bit like [github.com/hmans/beans](https://github.com/hmans/beans) and
[github.com/steveyegge/beads/](https://github.com/steveyegge/beads/) but much
simpler, providing only the core features of a task tracker.

## Why Pebbles?

When working with AI coding agents, you'll often have multiple tasks, bugs, and
features being worked on simultaneously. Pebbles provides:

- **Task tracking** - Create, read, update, and query tasks as JSON objects
- **Dependency management** - Link related work with `blocked-by` relationships
- **Status tracking** - Track progress through `new` → `in-progress` → `fixed`
  lifecycle
- **File-based storage** - All tasks stored as individual markdown files in
  `.pebbles/` directory (minimizes merge conflicts)

Pebbles is simple: It's less than 2k lines of code excluding tests and it has
only a handful of commands.

## Quick Start

Install pebbles:

```bash
go install go.yozora.eu/pebbles/cmd/peb@latest
```

Then run in your project directory:

```bash
peb init [--opencode] [--pi]
```

This creates a `.pebbles/` directory with a config file, optionally with an
opencode plugin and/or a pi agent extension. Edit the config file as needed.

## Opencode Integration

If you use opencode, install the opencode plugin, which provides:

- **Agent Instructions** to teach the agent how to use pebbles
- **MCP Server** with tools for each peb command

### Setup

Run `peb init --opencode` to install the plugin to
`.opencode/plugin/pebbles.ts`. This command does not override your main pebbles
config if it already exists.

### Automatic Updates

The plugin file in the project directory is automatically updated when running
any `peb` command if a newer version is available. This ensures that the plugin
stays in sync with the installed `peb` binary.

To disable the auto update of the opencode plugin, remove the "Version" string
from the first line of `.opencode/plugin/pebbles.ts`.

## Pi Integration

If you use [pi](https://github.com/earendil-works/pi), install the pi extension,
which provides:

- **Agent Instructions** (the pebbles "prime" prompt) injected into the system
  prompt on every turn
- **Tools** for each peb command (`peb_new`, `peb_read`, `peb_update`,
  `peb_query`, `peb_delete`) that the agent can call directly
- **`fix_peb`** to delegate fixing a single peb to an isolated **background**
  subagent in a throwaway jj worktree (non-blocking), plus **`fix_peb_list`** /
  **`fix_peb_kill`** to monitor and abort jobs (see
  [Delegating Fixes with `fix_peb`](#delegating-fixes-with-fix_peb))

### Setup

Run `peb init --pi` to install the extension to `.pi/extensions/pebbles.ts`.
This command does not override your main pebbles config if it already exists.

### Automatic Updates

The extension file in the project directory is automatically updated when
running any `peb` command if a newer version is available, just like the
opencode plugin. This ensures that the extension stays in sync with the
installed `peb` binary.

To disable the auto update of the pi extension, remove the "Version" string from
the first line of `.pi/extensions/pebbles.ts`.

### Delegating Fixes with `fix_peb`

The extension also registers a `fix_peb` tool that delegates fixing a single peb
to an **isolated background subagent**. `fix_peb` does **not** block the main
agent's turn — it returns immediately after starting the subagent, so the main
agent stays responsive and can keep chatting or launch more fixes.

When called, the tool:

1. Reads the peb.
2. Creates a temporary jj worktree off a configurable base revset
   (`jj workspace add -r <base> "<tmpdir>/<id>"`).
3. Optionally runs a worktree-initialization script (for example to install
   dependencies).
4. Spawns a subagent (no extensions) that fixes the peb and commits its work
   with `jj commit -m "<message>"`.
5. Returns immediately with a job id.
6. When the subagent finishes (**success or failure**), the main agent is
   notified via a message that includes the resulting jj change ids; the
   worktree is then forgotten and removed. (jj keeps the commits reachable after
   the workspace is forgotten, and the change ids are reported so the main agent
   can find them. Nothing is merged or pushed.)

The main agent may call `fix_peb` several times in one turn to fix multiple pebs
in parallel; each runs in its own isolated worktree and a process-wide semaphore
caps how many run at once (`maxParallel`).

Two companion tools let the main agent manage background fixes:

- **`fix_peb_list`** — list running and finished jobs with their status,
  worktree, subagent summary, and resulting commit change ids.
- **`fix_peb_kill`** — abort a running subagent by peb id (SIGTERM); its
  workspace is torn down and a failure notification is delivered.

Static configuration lives under the `pebbles.fixPeb` key of pi's settings, with
the project file overriding the global one and nested objects merged:

- global: `~/.pi/agent/settings.json`
- project: `.pi/settings.json`

```json
{
  "pebbles": {
    "fixPeb": {
      "baseRevset": "main",
      "worktreeInit": "cd \"$1\" && pnpm install",
      "subagentModel": "provider/model",
      "commitMessage": "fix: {title} ({id})",
      "timeoutMs": 1800000,
      "maxParallel": 4
    }
  }
}
```

All fields are optional; anything you omit keeps its default.

| Field           | Default           | Description                                                                                                                                                                                          |
| --------------- | ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `baseRevset`    | `"main"`          | jj revset the worktree is branched from.                                                                                                                                                             |
| `worktreeInit`  | _(none)_          | Shell script run after creating the worktree. It runs in the main repository's working directory and receives the worktree path as `$1` (for example `ln -s "$PWD/node_modules" "$1/node_modules"`). |
| `subagentModel` | _(current model)_ | `provider/id` for the subagent. When unset, the subagent uses the main agent's currently selected model.                                                                                             |
| `commitMessage` | `"<message>"`     | Commit-message template; `{id}` and `{title}` are substituted from the peb and sanitized to a single, shell-safe line.                                                                               |
| `timeoutMs`     | `1800000`         | Per-subagent timeout in milliseconds (30 minutes).                                                                                                                                                   |
| `maxParallel`   | `6`               | Maximum number of subagents running at once.                                                                                                                                                         |

## Using Pebbles with Coding Agents

### Opencode Agent Workflow

With the plugin installed, opencode agents use peb automatically to track their
work.

#### Example Session

```
User: Add a user profile page

Agent: I'll create a task for this and start working on it.
[Creates peb: "Add user profile page" with status "new"]
[Updates peb to "in-progress"]
[Implements the feature]
[Updates peb to "fixed"]

Task peb-xyz1 marked as fixed!
```

### Pi Agent Workflow

With the extension installed, pi agents use peb automatically to track their
work. The pebbles instructions are injected into the system prompt, and the
`peb_*` tools are available for the agent to call.

### Other Coding Agents

Pebbles works with any coding agent that supports running shell commands. Show
your agent the output of `peb prime` to teach it how to use pebbles.

## Command Reference

### Human Commands

#### `peb init [--opencode] [--pi]`

Initialize pebbles in current directory (creates `.pebbles/`). With `--opencode`
flag, also installs the opencode MCP plugin (creates
`.opencode/plugin/pebbles.ts`). With `--pi` flag, also installs the pi agent
extension (creates `.pi/extensions/pebbles.ts`).

#### `peb cleanup`

Delete all closed pebs (permanently removes pebs with status `fixed` or
`wont-fix`). Open pebs (status `new` or `in-progress`) are preserved.

#### `peb config`

Display the current pebbles configuration as JSON. This command is primarily
used by the opencode plugin to dynamically load configuration at runtime.

```bash
peb config
```

Output example:

```json
{
  "prefix": "peb",
  "id_length": 4
}
```

### AI Agent Commands

> **Note:** The following commands are designed for AI agents and use JSON for
> input/output. They are not intended for human use.

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

#### `peb delete <id> [<id> ...]`

Delete one or more tasks by ID (permanently removes them)

```bash
peb delete peb-ab12
peb delete peb-ab12 peb-cd34 peb-ef56
```

#### `peb prime [--mcp]`

Output agent instructions. With `--mcp` flag, outputs instructions formatted for
MCP server integration.

## Data Model

Each peb has:

- **id**: Unique identifier (e.g., `peb-ab12`, the prefix is customizable via
  the config)
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

Inspired by [github.com/hmans/beans](https://github.com/hmans/beans) and
[github.com/steveyegge/beads/](https://github.com/steveyegge/beads/)
