---
id: peb-xe8x
title: Add Blocking field to Peb data structure
type: epic
status: new
created: "2026-01-19T22:07:15+01:00"
changed: "2026-01-19T22:14:40+01:00"
blocked-by:
    - peb-0fqz
    - peb-7epm
    - peb-99nn
    - peb-bwgc
    - peb-3vzy
    - peb-wstm
    - peb-wpre
    - peb-4zns
    - peb-wzqx
---
Add a `Blocking []string` field to the Peb data structure that is the inverse of `BlockedBy`. If peb-A has `blocked-by: [peb-B]`, then peb-B should have `blocking: [peb-A]`.

All commands that modify pebs must ensure that the blocking/blocked-by entries are symmetrical. The query tool should allow querying for `blocking:peb-id`.

Implementation requirements:
- Persist `Blocking` to disk in YAML frontmatter
- Auto-repair blocking/blocked-by symmetry on save
- Allow users to set `blocking` directly via peb new/update (which updates blocked-by on the other peb)
- Handle all edge cases: self-reference, duplicates, invalid references, cycles, concurrent modifications

Subtasks:
- Add Blocking field to data model
- Implement symmetry enforcement logic
- Update new command to support blocking
- Update update command to support blocking  
- Update delete command to clean up relationships
- Update query command to support blocking filter
- Update validation and cycle detection
- Update MCP tools and documentation
- Add comprehensive test coverage