---
# pebbles-m0dp
title: Implement peb new command
status: completed
type: task
priority: normal
created_at: 2026-01-17T19:44:09Z
updated_at: 2026-01-17T22:06:13Z
parent: pebbles-6uma
blocking:
    - pebbles-988k
---

Implement the `peb new` command to create new pebs.

## Command Syntax

```
peb new
```

Reads JSON from stdin.

## JSON Input Format

Required fields: `title`, `content`

Optional fields: `type` (default: `bug`), `blocked-by` (array of peb IDs, or empty array [] to clear)

## Checklist

- [x] Create `internal/commands/new.go`
- [x] Implement new command that:
  - Reads JSON from stdin
  - Generates unique ID using configured prefix and length
  - Creates peb file with proper naming convention
  - Sets created/changed timestamps to current local time with timezone
  - Outputs: `Created new pebble $id in .pebbles/$filename`
- [x] Validate blocked-by references exist
  - Output error: `Error: Referenced pebble(s) not found: $id`
- [x] Register command in main.go
- [x] Write tests for new command