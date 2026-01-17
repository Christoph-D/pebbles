---
# pebbles-m0dp
title: Implement peb new command
status: todo
type: task
priority: normal
created_at: 2026-01-17T19:44:09Z
updated_at: 2026-01-17T19:44:50Z
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

- [ ] Create `internal/commands/new.go`
- [ ] Implement new command that:
  - Reads JSON from stdin
  - Generates unique ID using configured prefix and length
  - Creates peb file with proper naming convention
  - Sets created/changed timestamps to current local time with timezone
  - Outputs: `Created new pebble $id in .pebbles/$filename`
- [ ] Validate blocked-by references exist
  - Output error: `Error: Referenced pebble(s) not found: $id`
- [ ] Register command in main.go
- [ ] Write tests for new command