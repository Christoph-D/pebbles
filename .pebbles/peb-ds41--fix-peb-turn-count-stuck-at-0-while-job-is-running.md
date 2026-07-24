---
id: peb-ds41
title: fix_peb turn count stuck at 0 while job is running
type: bug
status: fixed
created: "2026-07-24T22:32:28+02:00"
changed: "2026-07-24T22:38:49+02:00"
---
## Problem

In `internal/config/data/pebbles-pi.ts`, the `fix_peb` tool tracks a per-job `turns` counter that is reported by `fix_peb_list`. However, it is always `0` while a job is running.

## Root cause

The turn count lives on the internal `SubagentResult` object (`result.turns++` at line ~267 inside `spawnSubagent`). The `FixJob` that `fix_peb_list` reads is initialized to `turns: 0` (line ~847) and only synced from the result object **once, on completion** (`job.turns = sub.turns` at line ~859).

Because `fix_peb` is explicitly a *background* tool whose prompt tells the caller to "use fix_peb_list to monitor jobs", the natural workflow is to poll `fix_peb_list` while jobs are still running — where `turns` is always `0`. The same staleness applies to `summary`, `model`, `stopReason`, and `errorMessage`.

(Verified: the `message_end` parsing itself works — a real multi-turn `pi --mode json` run counts `turns: 2`, and a completed `fix_peb` job produces a summary/commits proving assistant `message_end` events fired. So completed jobs are correct; only the running state is broken.)

## Fix

Propagate live progress from the stream to the `FixJob` as it arrives:

1. Add an optional `onProgress?: (result: SubagentResult) => void` callback to `spawnSubagent`'s options and invoke it at the end of the assistant `message_end` branch (after all the `result.*` mutations).
2. In the `fix_peb` `execute`, build the `FixJob` shell *before* spawning, pass an `onProgress` that mirrors `turns`/`summary`/`model`/`stopReason`/`errorMessage` onto the job, then assign `job.proc` after spawn.
3. Surface the turn count in the `fix_peb` completion notification (`notifyFixComplete`) so "turn reporting" is visible there too.

## Acceptance criteria

- [ ] `fix_peb_list` shows a non-zero `turns` value for a job while it is still running.
- [ ] Completion notification includes the turn count.
- [ ] Completed jobs still report the correct (unchanged) turn count.
- [ ] `peb` still builds (`make`) and the rendered extension has no TS errors.