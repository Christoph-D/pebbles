---
id: peb-ufvw
title: Move blocked-by field last and omit null values in peb query output
type: feature
status: fixed
created: "2026-01-18T12:44:22+01:00"
changed: "2026-01-18T12:52:53+01:00"
---
Change peb query so that:
1. blocked-by comes last in the output
2. blocked-by:null is omitted