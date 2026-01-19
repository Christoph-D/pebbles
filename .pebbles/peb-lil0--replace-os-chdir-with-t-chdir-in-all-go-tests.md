---
id: peb-lil0
title: Replace os.Chdir with t.Chdir in all Go tests
type: task
status: fixed
created: "2026-01-19T19:08:35+01:00"
changed: "2026-01-19T19:12:46+01:00"
---
Replace all usages of `os.Chdir` with `t.Chdir` in Go test files. This change:
- Removes need for manual defer cleanup
- Makes tests cleaner and less error-prone
- Ensures proper cleanup after each test

Files to update:
- internal/commands/delete_test.go
- internal/commands/init_test.go
- internal/commands/query_test.go
- internal/commands/update_test.go
- internal/commands/cleanup_test.go

Each test file has multiple tests using the pattern:
```go
origWd, _ := os.Getwd()
defer os.Chdir(origWd)
os.Chdir(testDir)
```

Should be replaced with:
```go
t.Chdir(testDir)
```