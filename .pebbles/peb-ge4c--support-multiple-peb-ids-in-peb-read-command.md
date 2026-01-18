---
id: peb-ge4c
title: Support multiple peb IDs in peb read command
type: feature
status: new
created: "2026-01-18T12:36:45+01:00"
changed: "2026-01-18T13:34:01+01:00"
---
The `peb read` command should accept an arbitrary number of peb IDs and read/display all of them. Currently it only accepts a single ID.

## Implementation Notes
- Update the `peb read` implementation and tests
- Update the MCP tool in `.opencode/plugin/pebbles-prime.ts` to support multiple IDs
- Update all documentation to reflect this change
- Ensure backward compatibility with single ID usage
