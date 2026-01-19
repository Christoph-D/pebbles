---
id: peb-wpre
title: Update validation and cycle detection
type: task
status: new
created: "2026-01-19T22:07:59+01:00"
changed: "2026-01-19T22:09:23+01:00"
---
Update validation in `internal/peb/validate.go`.

Requirements:
- Add `ValidateBlocking()` function (similar to ValidateBlockedBy)
- Update cycle detection to consider both blocking and blocked-by directions
- Setting blocking:peb-B on peb-A when peb-B has blocked-by:peb-A should detect a cycle