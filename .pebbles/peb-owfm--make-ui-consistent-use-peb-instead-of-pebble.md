---
id: peb-owfm
title: Make UI consistent - use "peb" instead of "pebble"
type: task
status: fixed
created: "2026-01-19T23:07:07+01:00"
changed: "2026-01-19T23:12:35+01:00"
---
Update the UI to be consistent:
- Change all references from "pebble" to "peb"
- Align the output of different commands for consistency

This requires:
1. Finding all instances of "pebble" in the codebase
2. Updating CLI command outputs
3. Updating help text and documentation
4. Ensuring consistent formatting across commands