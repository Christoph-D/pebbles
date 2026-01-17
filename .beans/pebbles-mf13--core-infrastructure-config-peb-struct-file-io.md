---
# pebbles-mf13
title: Core infrastructure - config, Peb struct, file I/O
status: completed
type: task
priority: normal
created_at: 2026-01-17T19:44:01Z
updated_at: 2026-01-17T19:44:50Z
parent: pebbles-6uma
blocking:
    - pebbles-bfa9
---

Implement the core infrastructure for the pebbles CLI tool.

## Checklist

- [x] Set up Go module (`go.mod`) with required dependencies:
  - `github.com/urfave/cli/v2`
  - `gopkg.in/yaml.v3`
  - `github.com/BurntSushi/toml`
- [x] Create `internal/config/config.go`:
  - Config struct with `Prefix` (default: "peb") and `IDLength` (default: 4)
  - `Load()` function that traverses upward to find `.pebbles/` directory
  - Parse `.pebbles/config.toml`
- [x] Create `internal/peb/peb.go`:
  - Peb struct with fields: ID, Title, Type, Status, Created, Changed, BlockedBy, Content
  - Type constants: bug, feature, epic, task
  - Status constants: new, in-progress, fixed, wont-fix
- [x] Create `internal/peb/id.go`:
  - ID generation: `$prefix-$random` where `$random` is N chars [0-9a-z]
  - Collision checking function
- [x] Create `internal/peb/file.go`:
  - File naming: `peb-$id--$title.md` (title: lowercase, spaces→dashes, remove non-alphanumeric)
  - Read/write functions for markdown with YAML frontmatter
  - Parse and serialize peb files
- [x] Create `internal/peb/validate.go`:
  - Cycle detection for blocked-by relationships (DFS)
  - Reference integrity validation (all IDs must exist)
- [x] Create `internal/store/store.go`:
  - Store struct that manages collection of pebs
  - Load all pebs from `.pebbles/` directory
  - Save individual peb files
  - Query/filter functions
- [x] Create `cmd/peb/main.go`:
  - Basic CLI app setup with urfave/cli
  - Root command structure