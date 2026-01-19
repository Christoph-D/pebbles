---
id: peb-7epm
title: Implement symmetry enforcement logic
type: task
status: wont-fix
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:25:05+01:00"
---
Implement symmetry enforcement logic to keep blocking/blocked-by relationships in sync.

Create helper functions in `internal/peb/sync.go`:
- `SyncBlockingRelationships()` - syncs both directions
- Handle blocked-by changes: add/remove from blocking lists
- Handle blocking changes: add/remove from blocked-by lists
- Return list of modified peb IDs that need to be saved