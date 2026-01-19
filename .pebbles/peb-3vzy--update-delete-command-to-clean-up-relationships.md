---
id: peb-3vzy
title: Update delete command to clean up relationships
type: task
status: new
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:09:23+01:00"
---
Update `internal/commands/delete.go` to clean up relationships before deletion.

Requirements:
- Before deleting, remove peb ID from blocking lists of all pebs in its blocked-by
- Remove peb ID from blocked-by lists of all pebs in its blocking
- Save all modified pebs before deletion