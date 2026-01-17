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
peb query [--fields <field,...>] [filters...]
```

## Output Format

Returns JSONL (JSON Lines) - one JSON object per line.

## Filters

- `status:<new|in-progress|fixed|wont-fix>` - Filter by status
- `type:<bug|feature|epic|task>` - Filter by type
- `blocked-by:<peb-id>` - Pebs that have this peb in their blocked-by list

Multiple filters use implicit AND.

## Fields

Default fields: `id`, `type`, `status`, `title`

## Checklist

- [ ] Create `internal/commands/query.go`
- [ ] Implement query command that:
  - No arguments: lists all pebs as JSONL
  - Parses filter arguments in `key:value` format
  - Applies filters with AND logic
  - Outputs each matching peb as JSON line with specified fields
- [ ] Implement `--fields` flag for customizing output fields
- [ ] Implement status filter
- [ ] Implement type filter
- [ ] Implement blocked-by filter (find pebs that have the given ID in their blocked-by list)
- [ ] Register command in main.go
- [ ] Write tests for query command with various filter combinations