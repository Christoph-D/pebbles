---
id: peb-wpre
title: Update validation and cycle detection
type: task
status: wont-fix
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:25:09+01:00"
---
Update validation in `internal/peb/validate.go`.

Requirements:
- Add `ValidateBlocking()` function (similar to ValidateBlockedBy)
- Update cycle detection to consider both blocking and blocked-by directions
- Setting blocking:peb-B on peb-A when peb-B has blocked-by:peb-A should detect a cycle