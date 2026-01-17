---
# pebbles-988k
title: Implement peb read command
status: completed
type: task
priority: normal
created_at: 2026-01-17T19:44:11Z
updated_at: 2026-01-17T22:31:05Z
parent: pebbles-6uma
blocking:
    - pebbles-m0yt
---

Implement the `peb read` command to display peb content as JSON.

## Command Syntax

```
peb read <peb-id>
```

## Checklist

- [x] Create `internal/commands/read.go`
- [x] Implement read command that:
  - Takes peb ID as argument
  - Loads the peb file
  - Outputs full peb data as JSON (including all fields: id, title, type, status, created, changed, blocked-by, content)
- [x] Handle error case: peb not found
- [x] Register command in main.go
- [x] Write tests for read command