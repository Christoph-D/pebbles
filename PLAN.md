# Pebbles - Task Tracking CLI Tool

A Go CLI tool called "pebbles" (binary: `peb`) for tracking tasks, bugs, features, and epics.

## Configuration

- Configured through `.pebbles/config.toml`
- At startup, traverses from current directory upwards until it finds `.pebbles/` directory
- Peb data stored in `.pebbles/` directory

### `.pebbles/config.toml` Schema

```toml
prefix = "peb"      # Prefix for peb IDs (default: "peb")
id_length = 4       # Length of random ID string (default: 4)
```

## Data Model

### Peb Structure

A "peb" can be a task, feature, bug, or epic with the following fields:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier: `$prefix-$random` where `$random` is N characters [0-9a-z] |
| `title` | Short description |
| `type` | One of: `bug` (default), `feature`, `epic`, `task` |
| `status` | One of: `new` (default), `in-progress`, `fixed`, `wont-fix` |
| `created` | Local timestamp in ISO format with timezone offset |
| `changed` | Local timestamp in ISO format with timezone offset (updated on any change) |
| `blocked-by` | List of peb IDs that block this peb (optional) |
| `content` | Markdown content |

### Relationships

- **Blocked-by**: A peb can be blocked by zero or more pebs. No cycles allowed.

## File Format

### File Location & Naming

- Stored in `.pebbles/` directory
- Filename pattern: `peb-$id--$title.md`
  - `$title`: lowercase, spaces replaced with dashes, non-alphanumeric characters removed

### Markdown Format

YAML frontmatter followed by markdown content. Empty fields are omitted.

**Minimal peb:**
```markdown
---
id: peb-f3zh
title: Fix the login system
type: bug
status: new
created: 2026-01-01T12:00:00-08:00
changed: 2026-01-01T12:00:00-08:00
---
Users can't log in if their name is "null".
```

**Peb with relationships:**
```markdown
---
id: peb-f3zh
title: Fix the login system
type: bug
status: new
created: 2026-01-01T12:00:00-08:00
changed: 2026-01-01T12:00:00-08:00
blocked-by:
  - peb-qwer
---
Users can't log in if their name is "null".
```

## CLI Commands

### `peb init`

Initialize a new pebbles project in the current directory. Creates `.pebbles/` directory with `config.toml`.

```
peb init
```

**Example:**
```
$ peb init
Initialized pebbles in .pebbles/

$ cat .pebbles/config.toml
# Pebbles configuration
prefix = "peb"
id_length = 4
```

**Behavior:**
- Creates `.pebbles/` directory with `config.toml`
- Fails if `.pebbles/` already exists in the current directory

```
$ peb init
Error: .pebbles/ already exists in current directory.
```

### `peb new`

Create a new peb.

```
peb new <title> <content> [--type <type>] [--blocked-by <peb-id,...>]
```

**Examples:**
```
$ peb new "Dependency" "Must be done first" --blocked-by peb-abc1
Created new pebble peb-mn34 in .pebbles/peb-mn34--dependency.md

$ peb new "Some title" "..." --blocked-by peb-nonexistent
Error: Blocked-by pebble(s) not found: peb-nonexistent
```

### `peb read`

Display a peb's full content.

```
peb <peb-id> read
```

**Example:**
```
$ peb peb-f3zh read
---
id: peb-f3zh
title: Fix the login system
type: bug
status: new
created: 2026-01-01T12:00:00-08:00
changed: 2026-01-01T12:00:00-08:00
blocked-by:
  - peb-qwer
---
Users can't log in if their name is "null".
```

### `peb set-status`

Update a peb's status.

```
peb <peb-id> set-status <new|in-progress|fixed|wont-fix>
```

**Example:**
```
$ peb peb-f3zh set-status in-progress
Marked pebble peb-f3zh "Fix the login system" as in-progress.
```

### `peb set-title`

Update a peb's title. This renames the file.

```
peb <peb-id> set-title <title>
```

**Example:**
```
$ peb peb-f3zh set-title "Fix the authentication system"
Updated title of peb-f3zh to "Fix the authentication system".
```

### `peb set-content`

Update a peb's content.

```
peb <peb-id> set-content <content>
```

**Example:**
```
$ peb peb-f3zh set-content "Updated description of the issue."
Updated content of peb-f3zh "Fix the login system".
```

### `peb set-type`

Update a peb's type.

```
peb <peb-id> set-type <bug|feature|epic|task>
```

**Example:**
```
$ peb peb-f3zh set-type feature
Updated type of peb-f3zh "Fix the login system" to feature.
```

### `peb set-blocking`

Update a peb's blocking list. Updates the "blocked-by" side of the relationship. Use empty string to clear.

```
peb <peb-id> set-blocking <peb-id,...|"">
```

**Example:**
```
$ peb peb-f3zh set-blocking peb-5n20,peb-drg8
Updated blocking list of peb-f3zh "Fix the login system".

$ peb peb-f3zh set-blocking ""
Cleared blocking list of peb-f3zh "Fix the login system".
```

**Behavior:**
1. Adds target to each blocked peb's `blocked-by` list
2. Removes target from any previously blocked peb's `blocked-by` list

### `peb set-blocked-by`

Update a peb's blocked-by list. Updates both sides of the relationship. Use empty string to clear.

```
peb <peb-id> set-blocked-by <peb-id,...|"">
```

**Example:**
```
$ peb peb-f3zh set-blocked-by peb-qwer,peb-asdf
Updated blocked-by list of peb-f3zh "Fix the login system".

$ peb peb-f3zh set-blocked-by ""
Cleared blocked-by list of peb-f3zh "Fix the login system".
```

**Behavior:**
1. Sets target peb's `blocked-by` field
2. Validates that all referenced peb IDs exist
3. Validates no cycles in the dependency graph

### `peb query`

Search and list pebs. No arguments lists all pebs. Multiple filters use implicit AND.

```
peb query [filters...]
```

**Filters:**
- `status:<new|in-progress|fixed|wont-fix>` - Filter by status
- `type:<bug|feature|epic|task>` - Filter by type
- `blocked-by:<peb-id>` - Pebs that are blocked by the given peb
- `blocking:<peb-id>` - Pebs that block the given peb

**Examples:**
```
$ peb query
peb-f3zh (bug,new) Fix the login system
peb-abc1 (bug,new) Some bug
peb-drg8 (bug,in-progress) Some other bug
peb-jwyp (feature,new) Some feature

$ peb query status:new
peb-f3zh (bug,new) Fix the login system
peb-abc1 (bug,new) Some bug
peb-jwyp (feature,new) Some feature

$ peb query status:new type:feature
peb-jwyp (feature,new) Some feature

$ peb query blocked-by:peb-f3zh
peb-5n20 (bug,new) Some bug
peb-drg8 (bug,in-progress) Some other bug

$ peb query blocked-by:peb-f3zh status:new
peb-5n20 (bug,new) Some bug

$ peb query blocking:peb-f3zh
No pebbles found.
```

## Project Structure

```
pebbles/
├── cmd/
│   └── peb/
│       └── main.go              # Entry point, CLI setup
├── internal/
│   ├── config/
│   │   └── config.go            # config.toml loading, traversal & init
│   ├── peb/
│   │   ├── peb.go               # Peb struct & core types
│   │   ├── id.go                # ID generation ($prefix-$random)
│   │   ├── file.go              # File naming, read/write markdown+YAML
│   │   └── validate.go          # Cycle detection, validation
│   ├── store/
│   │   └── store.go             # Load/save pebs, query index
│   └── commands/
│       ├── init.go              # peb init
│       ├── new.go               # peb new
│       ├── read.go              # peb read
│       ├── setters.go           # peb set-* commands
│       └── query.go             # peb query
├── go.mod
└── go.sum
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/urfave/cli/v2` | CLI framework |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing |
| `github.com/BurntSushi/toml` | Config file parsing |

## Validation Rules

1. **Cycle detection**: No cycles allowed in blocked-by relationships
2. **Reference integrity**: All referenced peb IDs must exist
3. **Unique IDs**: Check for collisions when generating new IDs

## Implementation Order

1. Core infrastructure - config loading, Peb struct, file I/O
2. `peb init` - initialize new project
3. `peb new` - create pebs with all flags
4. `peb read` - display peb content
5. `peb set-*` commands - all setters with bidirectional sync
6. `peb query` - filtering and listing
