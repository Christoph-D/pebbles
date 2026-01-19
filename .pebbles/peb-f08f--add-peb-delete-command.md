---
id: peb-f08f
title: Add peb delete command
type: feature
status: fixed
created: "2026-01-19T18:51:12+01:00"
changed: "2026-01-19T18:55:28+01:00"
---
Implement a new `peb delete` command that accepts one or more peb IDs and deletes the corresponding peb files from storage.

Implementation details:
- Created `internal/commands/delete.go` with DeleteCommand() function
- Added delete command to `cmd/peb/main.go`
- Created `internal/commands/delete_test.go` with comprehensive tests
- Updated MCP tools in `internal/commands/data/pebbles.ts` to include peb_delete tool
- Updated documentation in `internal/commands/data/prompt.md` and `internal/commands/data/prompt-mcp.md`
- Updated README.md to document the new command

The delete command:
- Accepts one or more peb IDs as arguments
- Pre-validates all pebs exist before any deletion
- Validates that no pebs reference the pebs being deleted via blocked-by
- Permanently removes peb files from storage
- Provides user-friendly output for each deleted peb
- Includes appropriate warnings about destructive nature of the command

Tests:
- TestDeleteCommand: Verifies single peb deletion
- TestDeleteCommandMultipleIDs: Verifies multiple peb deletion with output content validation
- TestDeleteCommandNotFound: Verifies error for non-existent peb
- TestDeleteCommandNoArgs: Verifies error for missing arguments
- TestDeleteCommandWithDependentPebs: Verifies deletion fails when peb has dependent pebs
- TestDeleteCommandPartialFailure: Verifies existing peb is not deleted when one peb in list is non-existent