---
id: peb-t24u
title: 'peb read prints BlockedBy: null when blocked-by is empty'
type: bug
status: fixed
created: "2026-01-18T01:01:30+01:00"
changed: "2026-01-18T01:44:28+01:00"
---
The "peb read" command prints "BlockedBy: null" when the blocked-by field is empty. It should instead omit this field from the output when there are no dependencies.