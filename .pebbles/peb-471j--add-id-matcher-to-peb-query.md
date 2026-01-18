---
id: peb-471j
title: 'Add id: matcher to peb query'
type: feature
status: fixed
created: "2026-01-18T13:59:36+01:00"
changed: "2026-01-18T14:09:16+01:00"
---
Add an id: matcher to the peb query command that supports:
- Single ID query: id:peb-xxxx
- Multiple ID query: id:(peb-xxxx|peb-yyyy)

This requires updating:
- The query command implementation in internal/commands/
- The MCP tool in .opencode/plugin/pebbles.ts
- All documentation/markdown files