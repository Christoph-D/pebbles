---
id: peb-2qlc
title: Implement lazy loading in store
type: feature
status: fixed
created: "2026-01-18T11:55:18+01:00"
changed: "2026-01-18T12:11:31+01:00"
---
Modify internal/store/store.go to use lazy loading:
1. Initially read only the directory to create a map ID -> filename
2. Whenever a peb is requested, either return it from the cache or load the file