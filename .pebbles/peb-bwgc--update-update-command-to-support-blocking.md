---
id: peb-bwgc
title: Update update command to support blocking
type: task
status: wont-fix
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:25:07+01:00"
---
Update `internal/commands/update.go` to support the `blocking` field.

Requirements:
- Add `Blocking *[]string` to `UpdateInput` struct
- Validate blocking IDs exist
- Check for cycles in both directions
- Sync removed blocking entries: remove peb ID from their blocked-by
- Sync added blocking entries: add peb ID to their blocked-by
- Sync blocked-by changes similarly
- Save all modified pebs