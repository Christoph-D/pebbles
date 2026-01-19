---
id: peb-99nn
title: Update new command to support blocking
type: task
status: new
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:09:23+01:00"
---
Update `internal/commands/new.go` to support the `blocking` field.

Requirements:
- Add `Blocking []string` to `NewInput` struct
- Validate blocking IDs exist
- After creating peb: add its ID to blocking lists of pebs it's blocked-by, and add its ID to blocked-by lists of pebs it's blocking
- Save all modified pebs