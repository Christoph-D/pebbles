---
id: peb-h9ow
title: peb init should be idempotent
type: feature
status: fixed
created: "2026-01-18T13:33:29+01:00"
changed: "2026-01-18T22:59:21+01:00"
---
Make peb init command idempotent - running it multiple times should not cause errors or duplicate data. Update all documentation to reflect this behavior.