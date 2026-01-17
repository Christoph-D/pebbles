---
# pebbles-bfa9
title: Implement peb init command
status: todo
type: task
priority: normal
created_at: 2026-01-17T19:44:04Z
updated_at: 2026-01-17T19:44:50Z
parent: pebbles-6uma
blocking:
    - pebbles-m0dp
---

Implement the `peb init` command to initialize a new pebbles project.

## Checklist

- [ ] Create `internal/commands/init.go`
- [ ] Implement init command that:
  - Creates `.pebbles/` directory in current directory
  - Creates `.pebbles/config.toml` with default values:
    ```toml
    # Pebbles configuration
    prefix = "peb"
    id_length = 4
    ```
  - Outputs: `Initialized pebbles in .pebbles/`
- [ ] Handle error case: `.pebbles/` already exists
  - Output: `Error: .pebbles/ already exists in current directory.`
- [ ] Register command in main.go
- [ ] Write tests for init command