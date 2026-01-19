---
id: peb-xe8x
title: Add Blocking field to Peb data structure
type: epic
status: wont-fix
created: "2026-01-19T22:07:15+01:00"
changed: "2026-01-19T22:25:04+01:00"
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

---

## Decision: Won't Fix

**Reason:** The persisted `blocking` field approach introduces significant complexity without proportional benefit.

### Problems with persisted approach:
1. **Data redundancy & integrity risk** - Storing the same relationship in two places creates risk of inconsistency if peb files are manually edited or corrupted
2. **Increased complexity** - Every command that modifies relationships (new, update, delete) must update multiple files atomically
3. **No atomic file operations** - Go/filesystem doesn't provide atomic multi-file writes; a crash mid-operation could leave data inconsistent
4. **Performance overhead** - Every write operation potentially requires loading and saving multiple pebs

### Alternative considered:
Computing `blocking` dynamically at read time was considered, but rejected because it requires loading every peb file to build the reverse lookup, making `peb_read` O(n) instead of O(k).

### Conclusion:
The current `blocked-by` field is sufficient. The `blocking` relationship can be derived when needed (e.g., `peb_query` with `blocked-by:peb-id` filter, or the existing `buildDependencyMap` logic in delete.go). The complexity of maintaining bidirectional persisted relationships outweighs the convenience benefit.