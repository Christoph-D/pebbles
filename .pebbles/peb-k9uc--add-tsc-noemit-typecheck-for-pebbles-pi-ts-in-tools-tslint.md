---
id: peb-k9uc
title: Add tsc --noEmit typecheck for pebbles-pi.ts in tools/tslint
type: task
status: fixed
created: "2026-07-25T00:01:54+02:00"
changed: "2026-07-25T00:10:29+02:00"
---
## Goal
Set up Option 2 (full `tsc --noEmit` typecheck) for the generated pi extension template at `internal/config/data/pebbles-pi.ts`, in a separate `tools/tslint/` directory.

## Why
`pebbles-pi.ts` is a Go text/template embedded into the `peb` binary and installed to `.pi/extensions/pebbles.ts`. It had no TypeScript tooling, so type errors against the pi SDK could slip in. (The only Go-template directive, `{{.Version}}`, is inside a comment, so no preprocessing is needed.)

## Done
- [x] Created `tools/tslint/package.json` (devDeps: typescript@5.9.3, @types/node@24.12.4, typebox@1.1.38, @earendil-works/pi-coding-agent@^0.82.0, pi-ai@^0.82.0, pi-tui@^0.82.0).
- [x] Created `tools/tslint/tsconfig.json` (noEmit, strict, skipLibCheck, module/moduleResolution esnext/bundler, `paths` mapping to `./node_modules` so the source file outside the dir resolves imports).
- [x] Generated + committed `tools/tslint/package-lock.json`.
- [x] `.gitignore` already has bare `node_modules` (covers `tools/tslint/node_modules`).
- [x] Added `lint-ts` Makefile target (incremental: only `npm ci` when deps change; runs `tsc --noEmit`).
- [x] Wired into `.github/workflows/test.yml` (setup-node@v4, node 24, npm cache, `make lint-ts`).
- [x] Documented `lint-ts` in `AGENTS.md`.

## Real type errors the linter caught & fixed in `pebbles-pi.ts` (baseline green)
1. `AgentToolResult<T>` requires `details` in pi 0.82.0 — added `details: undefined` to the 6 tool results (peb_new/read/update/query/delete + fix_peb_list/kill) and to the `onUpdate` emit in fix_peb.
2. `tryKill` declared `{ kill: (sig?: string) => boolean }` which is incompatible with node's real `ChildProcess.kill(signal?: number | NodeJS.Signals)` — corrected the helper's signature.

## Verification
- `make lint-ts` passes (fresh + incremental), exit 0.
- `make test` passes (all Go packages ok).
- `node_modules` correctly ignored by jj.