---
id: peb-v6nh
title: 'peb pi tools: toggle output with ctrl+o'
type: feature
status: fixed
created: "2026-07-24T20:12:25+02:00"
changed: "2026-07-24T20:12:30+02:00"
---
The peb tools registered by the pi extension (`pebbles-pi.ts`) did not honor the global ctrl+o (`app.tools.expand`) toggle. Because they registered no `renderResult`, the TUI used its fallback renderer which ignores the `expanded` flag and always shows full output — unlike pi's built-in tools (bash/read/edit/...) which collapse to a compact summary and expand on ctrl+o.

## Fix
In `internal/config/data/pebbles-pi.ts` (the embedded Go text/template rendered to `.pi/extensions/pebbles.ts`):
- Added a shared `renderPebResult` helper (plus `summarizePebOutput`, `tryParsePebJson`, `truncateForSummary`) that respects the `{ expanded }` render option:
  - collapsed → compact one-line summary (JSON-aware: `peb-xxxx: <title>` for read, `N pebs: ...` for query; first meaningful line otherwise) + `(N lines)` hint when multi-line
  - expanded → full text output
  - error styling for `Error:` output; `running…` for empty partial output
- Wired `renderResult: renderPebResult` into all six tools: `peb_new`, `peb_read`, `peb_update`, `peb_query`, `peb_delete`, `fix_peb`.
- Added imports: `Text` from `@earendil-works/pi-tui`, `Theme` type from `@earendil-works/pi-coding-agent`.
- Bumped `data/pebbles-pi.ts.version` (epoch → now) so `MaybeUpdatePlugin` rewrites already-installed `.pi/extensions/pebbles.ts` on next `peb` run.

## Verification
- `go build ./...` and `go test ./...` pass.
- Typechecked the rendered plugin against pi's real types (`@earendil-works/pi-coding-agent` + `pi-tui`): the only errors are 9 pre-existing version-skew artifacts (the `tryKill` node `ChildProcess` typing and `execute` returning `{content}` without the newer-required `details` field); none involve any added render symbol.