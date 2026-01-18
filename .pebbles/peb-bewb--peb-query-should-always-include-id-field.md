---
id: peb-bewb
title: peb query should always include id field
type: feature
status: new
created: "2026-01-18T12:37:20+01:00"
changed: "2026-01-18T12:37:20+01:00"
---
The id field should always be included in peb query output, regardless of whether --fields is specified or not. Currently, users can accidentally exclude the id field, making it difficult to identify or act on the results.