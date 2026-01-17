---
# pebbles-m0yt
title: Implement peb update command
status: todo
type: task
priority: normal
created_at: 2026-01-17T19:44:18Z
updated_at: 2026-01-17T19:44:50Z
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

- [ ] Create `internal/commands/update.go`
- [ ] Implement update command that:
  - Takes peb ID and JSON input (argument or stdin)
  - Updates specified fields (title, content, type, status, blocked-by)
  - Updates changed timestamp
  - Handles title updates with file rename
- [ ] For status updates, output: `Updated status of $id.`
- [ ] For title updates, output: `Updated title of $id.`
- [ ] For content updates, output: `Updated content of $id.`
- [ ] For type updates, output: `Updated type of $id.`
- [ ] For blocked-by updates:
  - Validate all referenced peb IDs exist
  - Validate no cycles would be created
  - Output: `Updated blocked-by list of $id.` or `Cleared blocked-by list of $id.`
- [ ] Register command in main.go
- [ ] Write tests for update command