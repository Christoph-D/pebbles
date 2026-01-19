---
id: peb-wstm
title: Update query command to support blocking filter
type: task
status: wont-fix
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:25:08+01:00"
---
Update `internal/commands/query.go` to support blocking filter and field.

Requirements:
- Add `blocking:peb-id` filter in `parseFilters`
- Add `"blocking"` to valid fields list in `parseFields`
- Add `"blocking"` to default fields
- Handle blocking field in `buildOutput`