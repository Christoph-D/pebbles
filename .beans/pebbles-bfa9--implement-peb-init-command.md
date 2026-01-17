---
# pebbles-bfa9
title: Implement peb init command
status: completed
type: task
priority: normal
created_at: 2026-01-17T19:44:04Z
updated_at: 2026-01-17T22:25:04Z
parent: pebbles-6uma
blocking:
    - pebbles-m0dp
---

Implement the `peb init` command to initialize a new pebbles project.

## Checklist

- [x] Create `internal/commands/init.go`
- [x] Implement init command that:
  - Creates `.pebbles/` directory in current directory
  - Creates `.pebbles/config.toml` with default values:
    ```toml
    # Pebbles configuration
    prefix = "peb"
    id_length = 4
    ```
  - Outputs: `Initialized pebbles in .pebbles/`
- [x] Handle error case: `.pebbles/` already exists
  - Output: `Error: .pebbles/ already exists in current directory.`
- [x] Register command in main.go
- [x] Write tests for init command