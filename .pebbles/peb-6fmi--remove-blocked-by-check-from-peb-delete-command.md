---
id: peb-6fmi
title: Remove blocked-by check from peb delete command
type: task
status: fixed
created: "2026-01-19T22:49:58+01:00"
changed: "2026-01-19T22:53:15+01:00"
---
Modify the peb delete command to delete arbitrary pebs without checking if they are blocked-by other pebs. Currently the delete command prevents deletion of pebs that are blocking other pebs.