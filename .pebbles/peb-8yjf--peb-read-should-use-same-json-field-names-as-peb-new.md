---
id: peb-8yjf
title: peb read should use same JSON field names as peb new
type: bug
status: fixed
created: "2026-01-18T01:03:28+01:00"
changed: "2026-01-18T01:06:17+01:00"
---
The "peb read" command capitalizes field names and prints "BlockedBy" instead of "blocked-by". JSON field names should be consistent with what "peb new" expects (e.g., "blocked-by", "created", "changed", etc.)