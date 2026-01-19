---
id: peb-wzqx
title: Add comprehensive test coverage
type: task
status: wont-fix
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:25:10+01:00"
---
Add comprehensive test coverage for blocking functionality.

Test files to update:
- `internal/commands/new_test.go` - test blocking on create, symmetry
- `internal/commands/update_test.go` - test blocking on update, bidirectional sync
- `internal/commands/delete_test.go` - test cleanup of blocking on delete
- `internal/commands/query_test.go` - test blocking filter
- Add validation tests for blocking