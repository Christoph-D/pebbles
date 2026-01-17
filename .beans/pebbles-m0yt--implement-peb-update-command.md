---
# pebbles-m0yt
title: Implement peb update command
status: completed
type: task
priority: normal
created_at: 2026-01-17T19:44:18Z
updated_at: 2026-01-17T22:50:05Z
parent: pebbles-6uma
blocking:
    - pebbles-1upa
---

Implement the `peb update` command to update peb fields.

## Command Syntax

```
peb update <peb-id> <JSON>
peb update <peb-id> < update.json
```

Takes a JSON object containing the fields to update (as argument or from stdin). Omitted fields remain unchanged.

Optional fields: `title`, `content`, `type`, `status`, `blocked-by`

## Checklist

- [x] Create `internal/commands/update.go`
- [x] Implement update command that:
  - Takes peb ID and JSON input (argument or stdin)
  - Updates specified fields (title, content, type, status, blocked-by)
  - Updates changed timestamp
  - Handles title updates with file rename
- [x] For status updates, output: `Updated status of $id.`
- [x] For title updates, output: `Updated title of $id.`
- [x] For content updates, output: `Updated content of $id.`
- [x] For type updates, output: `Updated type of $id.`
- [x] For blocked-by updates:
  - Validate all referenced peb IDs exist
  - Validate no cycles would be created
  - Output: `Updated blocked-by list of $id.` or `Cleared blocked-by list of $id.`
- [x] Register command in main.go
- [x] Write tests for update command