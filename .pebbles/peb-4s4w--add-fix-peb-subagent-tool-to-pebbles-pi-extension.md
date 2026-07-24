---
id: peb-4s4w
title: Add fix_peb subagent tool to pebbles-pi extension
type: feature
status: fixed
created: "2026-07-24T19:15:21+02:00"
changed: "2026-07-24T19:37:34+02:00"
---
## Goal
Add a `fix_peb` tool to the auto-generated pi extension that delegates fixing a single peb to an isolated subagent in a throwaway jj worktree.

## Done
- Extended `internal/config/data/pebbles-pi.ts` with:
  - `fix_peb` tool (params: `peb_id`, optional `extra_prompt`).
  - Module helpers: `loadFixPebConfig` (`.pi/fix-peb.json` merged over `~/.pi/agent/fix-peb.json`), `Semaphore` (caps `max_parallel`), `buildFixPrompt`, `runSubagent` (spawns `pi --mode json -p --no-session --no-extensions --no-context-files --approve [--model X]`).
  - Lifecycle: read peb → `jj workspace add -r <base> "<tmp>/<id>"` → optional `worktree_init` (`sh -c "$script" sh "<worktree>"`, cwd=main repo) → spawn subagent → capture new change ids via `<anchor>..@-` → `jj workspace forget` + `rm -rf <tmp>` in `finally`.
  - Model fallback: config `subagent_model` → else main agent's `ctx.model`.
  - All jj/subagent calls async (`pi.exec`/`spawn`) so parallel calls don't block each other.
- Updated README "Pi Integration" with `fix_peb` + config table.
- Bumped pi extension version; regenerated `.pi/extensions/pebbles.ts`.

## Commits
- vtzwtzsr "Add fix_peb subagent tool to pi extension" (template + peb + README)
- wkysuuro "Bump pi extension version for fix_peb" (version file + deployed ext)

## Verified
- Extension loads in pi; `fix_peb` registered (diagnostic dump of tools).
- jj workspace add/forget + temp dir lifecycle works.
- Change-id capture revset (`<anchor>..@-`) returns 1..N commits correctly.
- `worktree_init` semantics: `$1`=worktree, cwd=main repo.
- `go test ./internal/...` passes.

## Note
The mandatory `@code-reviewer` agent is not configured in this environment (no agents/skills dirs); performed a manual self-review instead. Known limitation: parallel `fix_peb` calls targeting the *same* peb id collide on the workspace name (clean error + cleanup); different pebs run independently.