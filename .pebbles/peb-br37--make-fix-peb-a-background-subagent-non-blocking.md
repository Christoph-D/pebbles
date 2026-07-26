---
id: peb-br37
title: Make fix_peb a background subagent (non-blocking)
type: feature
status: fixed
created: "2026-07-24T21:24:22+02:00"
changed: "2026-07-26T16:34:00+02:00"
---
## Goal

Rework the `fix_peb` pi extension tool (in `internal/config/data/pebbles-pi.ts`) so the subagent runs in the **background** instead of blocking the main agent's turn. The main agent must stay responsive to the user while fixes run.

## Requirements

1. `fix_peb` returns **immediately** after spawning the subagent (setup: read peb, create jj worktree, optional init, spawn `pi`). It must NOT `await` the subagent.
2. On subagent completion (**success or failure only** — never per text chunk), push a single notification to the main agent via `pi.sendMessage(..., { triggerTurn: true, deliverAs: "followUp" })`, including the new commit change ids and the subagent's summary (or error reason).
3. Keep existing teardown behavior: auto-commit + `jj workspace forget` + remove temp dir. jj keeps commits reachable after forget, and the change ids are reported to the main agent.
4. Add two tools:
   - `fix_peb_list` — list running + finished jobs (status, worktree, summary, change ids).
   - `fix_peb_kill` — SIGTERM a running subagent by peb id; results in a failure notification.
5. Register an idempotent `session_shutdown` handler that kills still-running subagents and tears down their workspaces. Children do **not** need to survive `/reload` or pi restart (no `detached`/`unref`, no `appendEntry` persistence).
6. Keep the `maxParallel` semaphore, now bounding live background jobs; acquire on start, release in the completion handler.

## Out of scope
- Surviving reload/restart / reattaching to child stdio.
- Merging/importing commits into the main repo (main agent handles via reported change ids).

## Acceptance
- `make` builds; `make test` passes.
- README `fix_peb` section updated to describe background semantics + new tools.
- Extension `.version` bumped so installed extensions update.