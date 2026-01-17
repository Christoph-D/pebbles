---
# pebbles-1upa
title: Implement peb query command
status: todo
type: task
priority: normal
created_at: 2026-01-17T19:44:23Z
updated_at: 2026-01-17T19:44:43Z
parent: pebbles-6uma
---

Implement the `peb query` command for searching and listing pebs.

## Command Syntax

```
peb query [filters...]
```

## Filters

- `status:<new|in-progress|fixed|wont-fix>` - Filter by status
- `type:<bug|feature|epic|task>` - Filter by type
- `blocked-by:<peb-id>` - Pebs that are blocked by the given peb
- `blocking:<peb-id>` - Pebs that block the given peb

Multiple filters use implicit AND.

## Checklist

- [ ] Create `internal/commands/query.go`
- [ ] Implement query command that:
  - No arguments: lists all pebs
  - Parses filter arguments in `key:value` format
  - Applies filters with AND logic
  - Outputs each peb as: `$id ($type,$status) $title`
  - Outputs `No pebbles found.` if no matches
- [ ] Implement status filter
- [ ] Implement type filter
- [ ] Implement blocked-by filter (find pebs that have the given ID in their blocked-by list)
- [ ] Implement blocking filter (find pebs whose ID appears in other pebs' blocked-by lists)
- [ ] Register command in main.go
- [ ] Write tests for query command with various filter combinations