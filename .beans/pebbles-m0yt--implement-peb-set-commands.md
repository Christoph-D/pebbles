---
# pebbles-m0yt
title: Implement peb set-* commands
status: todo
type: task
priority: normal
created_at: 2026-01-17T19:44:18Z
updated_at: 2026-01-17T19:44:50Z
parent: pebbles-6uma
blocking:
    - pebbles-1upa
---

Implement all `peb set-*` commands for updating peb fields.

## Commands

- `peb <peb-id> set-status <new|in-progress|fixed|wont-fix>`
- `peb <peb-id> set-title <title>`
- `peb <peb-id> set-content <content>`
- `peb <peb-id> set-type <bug|feature|epic|task>`
- `peb <peb-id> set-blocking <peb-id,...|"">`
- `peb <peb-id> set-blocked-by <peb-id,...|"">`

## Checklist

- [ ] Create `internal/commands/setters.go`
- [ ] Implement `set-status`:
  - Validate status value
  - Update status field
  - Update changed timestamp
  - Output: `Marked pebble $id "$title" as $status.`
- [ ] Implement `set-title`:
  - Update title field
  - Rename file to match new title
  - Update changed timestamp
  - Output: `Updated title of $id to "$title".`
- [ ] Implement `set-content`:
  - Update content (markdown body)
  - Update changed timestamp
  - Output: `Updated content of $id "$title".`
- [ ] Implement `set-type`:
  - Validate type value
  - Update type field
  - Update changed timestamp
  - Output: `Updated type of $id "$title" to $type.`
- [ ] Implement `set-blocking`:
  - Update blocked-by lists on target pebs (bidirectional sync)
  - Empty string clears the list
  - Update changed timestamp
  - Output: `Updated blocking list of $id "$title".` or `Cleared blocking list of $id "$title".`
- [ ] Implement `set-blocked-by`:
  - Validate all referenced peb IDs exist
  - Validate no cycles would be created
  - Update blocked-by field
  - Update changed timestamp
  - Output: `Updated blocked-by list of $id "$title".` or `Cleared blocked-by list of $id "$title".`
- [ ] Register all commands in main.go
- [ ] Write tests for all setter commands