---
id: peb-ud86
title: Clean blocked-by list in Save
type: task
status: fixed
created: "2026-01-19T22:31:15+01:00"
changed: "2026-01-19T22:35:11+01:00"
---
Modify the Save function in internal/store/store.go to clean the peb's blocked-by list before saving by removing non-existent entries. Must not change the argument.