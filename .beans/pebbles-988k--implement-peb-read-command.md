---
# pebbles-988k
title: Implement peb read command
status: todo
type: task
priority: normal
created_at: 2026-01-17T19:44:11Z
updated_at: 2026-01-17T19:44:50Z
parent: pebbles-6uma
blocking:
    - pebbles-m0yt
---

Implement the `peb read` command to display peb content.

## Command Syntax

```
peb <peb-id> read
```

## Checklist

- [ ] Create `internal/commands/read.go`
- [ ] Implement read command that:
  - Takes peb ID as argument
  - Loads the peb file
  - Outputs full markdown content with YAML frontmatter
- [ ] Handle error case: peb not found
- [ ] Register command in main.go
- [ ] Write tests for read command