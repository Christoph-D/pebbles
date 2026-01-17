# Pebbles - Task Tracking CLI Tool

A an agent-first Go CLI tool called "pebbles" (binary: `peb`) for tracking tasks, bugs, features, and epics. It is designed to be used by AI agents.

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

Create a new peb. Reads JSON from stdin.

```
peb new
```

Required fields: `title`, `content`.

Optional fields: `type` (default: `bug`), `blocked-by` (array of peb IDs, or empty array [] to clear).

**Examples:**
```
$ peb new <<'EOF'
{"title":"Dependent","content":"Must be done after peb-abc1","blocked-by":["peb-abc1"]}
EOF
Created new pebble peb-mn34 in .pebbles/peb-mn34--dependent.md

$ peb new <<'EOF'
{"title":"Some title","content":"...","blocked-by":"peb-nonexistent"}
EOF
Error: Referenced pebble(s) not found: peb-nonexistent
```

### `peb read`

Display a peb's full content as JSON.

```
peb read <peb-id>
```

**Example:**
```
$ peb read peb-f3zh
{
  "id": "peb-f3zh",
  "title": "Fix the login system",
  "type": "bug",
  "status": "new",
  "created": "2026-01-01T12:00:00-08:00",
  "changed": "2026-01-01T12:00:00-08:00",
  "blocked-by": ["peb-qwer"],
  "content": "Users can't log in if their name is \"null\"."
}
```

### `peb update`

Update a peb's fields. Takes a JSON object containing the fields to update (as argument or from stdin). Omitted fields remain unchanged.

```
peb update <peb-id> '{"title":"...","content":"...","type":"bug","status":"in-progress","blocked-by":["peb-id",...]}'
peb update <peb-id> < update.json
```

Optional fields: `title`, `content`, `type`, `status`, `blocked-by`.

**Examples:**
```
$ peb update peb-f3zh '{"status":"in-progress"}'
Updated status of peb-f3zh.

$ peb update peb-f3zh '{"title":"Fix the authentication system"}'
Updated title of peb-f3zh.

$ peb update peb-f3zh <<'EOF'
{"content":"Updated description with \"quotes\" and $special chars."}
EOF
Updated content of peb-f3zh.

$ peb update peb-f3zh '{"type":"feature"}'
Updated type of peb-f3zh.

$ peb update peb-f3zh '{"blocked-by":["peb-5n20","peb-drg8"]}'
Updated blocked-by list of peb-f3zh.

$ peb update peb-f3zh '{"blocked-by":[]}'
Cleared blocked-by list of peb-f3zh.
```

**Behavior for blocked-by field:**
1. Validates that all referenced peb IDs exist
2. Validates no cycles in the dependency graph
3. Sets the target peb's blocked-by list to the given peb IDs

### `peb query`

Search and list pebs. Returns JSONL (JSON Lines) - one JSON object per line. No arguments lists all pebs. Multiple filters use implicit AND.

**Default fields:** id, type, status, title

```
peb query [--fields <field,...>] [filters...]
```

**Filters:**
- `status:<new|in-progress|fixed|wont-fix>` - Filter by status
- `type:<bug|feature|epic|task>` - Filter by type
- `blocked-by:<peb-id>` - Pebs that have this peb in their blocked-by list

**Examples:**
```
$ peb query
{"id":"peb-f3zh","type":"bug","status":"new","title":"Fix the login system"}
{"id":"peb-abc1","type":"bug","status":"new","title":"Some bug"}
{"id":"peb-drg8","type":"bug","status":"in-progress","title":"Some other bug"}
{"id":"peb-jwyp","type":"feature","status":"new","title":"Some feature"}

$ peb query --fields id
{"id":"peb-f3zh"}
{"id":"peb-abc1"}
{"id":"peb-drg8"}
{"id":"peb-jwyp"}

$ peb query status:new
{"id":"peb-f3zh","type":"bug","status":"new","title":"Fix the login system"}
{"id":"peb-abc1","type":"bug","status":"new","title":"Some bug"}
{"id":"peb-jwyp","type":"feature","status":"new","title":"Some feature"}

$ peb query status:new type:feature
{"id":"peb-jwyp","type":"feature","status":"new","title":"Some feature"}

$ peb query blocked-by:peb-f3zh
{"id":"peb-5n20","type":"bug","status":"new","title":"Some bug"}
{"id":"peb-drg8","type":"bug","status":"in-progress","title":"Some other bug"}

$ peb query blocked-by:peb-f3zh status:new
{"id":"peb-5n20","type":"bug","status":"new","title":"Some bug"}
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
│       ├── update.go            # peb update
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
3. `peb new` - create pebs from JSON
4. `peb read` - display peb content
5. `peb update` - update pebs from JSON with bidirectional sync
6. `peb query` - filtering and listing
