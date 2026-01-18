---
id: peb-hc3c
title: Add GitHub Actions workflow to run tests
type: task
status: fixed
created: "2026-01-18T23:52:47+01:00"
changed: "2026-01-18T23:53:22+01:00"
---
Create a GitHub Actions workflow that runs `make test` on push and pull requests.

The workflow should:
- Run on push to main branch and on pull requests
- Use a recent version of Go
- Run `make test` to execute all tests
- Set up the Go environment properly