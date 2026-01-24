---
id: peb-9n4o
title: Call MaybeUpdatePlugin() in all peb subcommands
type: feature
status: fixed
created: "2026-01-24T18:30:40+01:00"
changed: "2026-01-24T18:31:42+01:00"
---
Add config.MaybeUpdatePlugin() calls to all peb subcommands (init, new, read, update, delete, query, cleanup, prime) to ensure the opencode plugin is kept up-to-date when commands are run.