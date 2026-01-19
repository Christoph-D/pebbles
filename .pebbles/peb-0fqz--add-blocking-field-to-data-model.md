---
id: peb-0fqz
title: Add Blocking field to data model
type: task
status: wont-fix
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:25:05+01:00"
---
Add `Blocking []string` field to both the `Peb` struct and `PebJSON` struct in `internal/peb/peb.go`.

Requirements:
- Add `Blocking []string` with YAML tag `yaml:"blocking,omitempty"` and JSON tag `json:"blocking,omitempty"`
- Add to `Peb` struct
- Add to `PebJSON` struct for query output
- Follow existing code style and patterns