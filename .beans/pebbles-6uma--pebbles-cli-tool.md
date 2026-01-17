---
# pebbles-6uma
title: Pebbles CLI Tool
status: completed
type: epic
created_at: 2026-01-17T19:43:36Z
updated_at: 2026-01-17T23:05:00Z
---

Implement a Go CLI tool called "pebbles" (binary: `peb`) for tracking tasks, bugs, features, and epics.

## Overview

- Configured through `.pebbles/config.toml`
- At startup, traverses from current directory upwards until it finds `.pebbles/` directory
- Peb data stored as markdown files with YAML frontmatter in `.pebbles/` directory

## Key Features

- Create and manage pebs (tasks, bugs, features, epics)
- Track blocked-by relationships with cycle detection
- Query and filter pebs by status, type, and relationships

## Dependencies

- `github.com/urfave/cli/v2` - CLI framework
- `gopkg.in/yaml.v3` - YAML frontmatter parsing
- `github.com/BurntSushi/toml` - Config file parsing

## Implementation Order

See child beans for detailed implementation tasks.