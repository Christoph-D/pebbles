---
id: peb-4s4w
title: Add fix_peb subagent tool to pebbles-pi extension
type: feature
status: in-progress
created: "2026-07-24T19:15:21+02:00"
changed: "2026-07-24T19:15:29+02:00"
---
## Goal
Add a `fix_peb` tool to the auto-generated pi extension (`internal/config/data/pebbles-pi.ts`) that lets the main agent delegate fixing a single peb to an isolated subagent running in a throwaway jj worktree.

## Behavior
1. Read the peb (`peb read <id>`).
2. Create a temp dir + a jj workspace branched off a configurable base revset (`jj workspace add -r <base> "<tmpdir>/<id>"`).
3. Optionally run a worktree-init shell script (cwd = main repo, `$1` = worktree path).
4. Spawn a subagent (`pi --mode json -p --no-session --no-extensions --no-context-files --approve [--model X] <prompt>`, cwd = worktree) instructed to fix the peb and `jj commit` its work.
5. Capture the subagent's result (exit code, final summary, usage) and the new jj change IDs.
6. Always tear down: `jj workspace forget <name>` and delete the temp dir (try/finally, runs on success/error/abort).
7. Report success/failure to the main agent. The main agent does nothing else with the commits.

## Tool contract
- `peb_id` (string, required)
- `extra_prompt` (string, optional)

Parallelism: multiple `fix_peb` calls in one turn run concurrently (pi runs tool calls in parallel). A process-local semaphore enforces `max_parallel` from config. Each call uses its own temp dir + unique workspace.

## Static config (`.pi/fix-peb.json`, project-local; merged over `~/.pi/agent/fix-peb.json`)
- `base_revset` (default `"main"`)
- `worktree_init` (optional shell script; receives worktree path as `$1`, runs in main repo)
- `subagent_model` (optional `provider/id`; falls back to the main agent's current model via `ctx.model`)
- `commit_message` (default `"fix: {title} ({id})"`; `{id}`/`{title}` substituted from the peb, sanitized to one line)
- `timeout_ms` (default 1800000)
- `max_parallel` (default 4)

## Non-goals
- Merging/abandoning the subagent's commits (they remain in the shared jj repo, findable by change id).
- Full streaming render machinery (we capture the subagent's final assistant message + usage only).

## Files
- `internal/config/data/pebbles-pi.ts` (the Go text/template source)
- Regenerate deployed `.pi/extensions/pebbles.ts` via `peb prime` after version bump.

## Verification
- Extension loads in pi without errors (`pi -p` smoke test).
- Manual jj workspace add/forget + temp dir flow works.