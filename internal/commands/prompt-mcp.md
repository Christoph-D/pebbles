# Using peb (Pebbles Task Tracker)

You are working with "peb" (Pebbles), an agent-first tool for tracking tasks, bugs, features, and epics.

## ⚠️ CRITICAL REQUIREMENT

**ALL NON-TRIVIAL WORK MUST BE TRACKED AS A PEB**

Before doing ANY non-trivial task, bug fix, feature, or code change:
1. Create a peb to track it (for complex work, create an epic with subtasks)
2. Update its status as you work
3. Mark it as fixed when complete

Trivial work (simple, single-step tasks that can be completed in < 3 minutes) can be done without a peb.

**Rules:**

1. Every non-trivial task, bug, feature, or code change must be tracked as a peb
2. Create a peb before starting the work
3. Update peb status throughout the lifecycle
4. Do not mark pebs as `fixed` until all dependencies (`blocked-by`) are also `fixed`
5. Use `blocked-by` to establish clear dependencies between related work
6. For complex work, create an `epic` peb that blocks smaller task pebs (epic remains `in-progress` until all tasks are `fixed`)

## Core Concepts

**Peb Structure:** A "peb" represents a task/bug/feature/epic with these fields:

- `id`: Unique identifier (format: `{{.PebbleIDPattern}}` where `{{.PebbleIDSuffix}}` is a random ID)
- `title`: Short description
- `type`: One of: `bug`, `feature`, `epic`, `task`
- `status`: One of: `new`, `in-progress`, `fixed`, `wont-fix`
- `created`/`changed`: timestamps
- `blocked-by`: List of peb IDs that block this peb (dependency tracking)
- `content`: Markdown description

Terminology:

- Pebs with status `new` or `in-progress` are "open". Query with `peb_query` using filters.
- Pebs with status `fixed` or `wont-fix` are "closed". Query with `peb_query` using filters.

## Best Practices

**Before starting work:**

1. Use `peb_query` with `filters: ["status:open"]` to find work
2. Use `peb_read` with the peb ID(s) to understand requirements

**While working:**

1. Use `peb_update` to mark pebs as `in-progress` when starting work
2. Use `peb_update` to update the peb's content if requirements change
3. Use `peb_update` to mark as `fixed` when completed

**Tracking dependencies:**

- Use `blocked_by` in `peb_new` when work depends on another peb
- Use `peb_read` to find a peb's dependencies (must be completed before this peb can be marked as fixed)
- Use `peb_query` with `filters: ["blocked-by:<id>"]` to find pebs blocked by a specific task

## Example Workflow

```
# 1. Find a bug to work on
Call peb_query with:
- filters: ["status:open", "type:bug"]

# 2. Read details
Call peb_read with:
- id: "{{.PebbleIDPattern}}"

# 3. Mark as in-progress
Call peb_update with:
- id: "{{.PebbleIDPattern}}"
- data: '{"status":"in-progress"}'

# 4. Do the work...

# 5. Mark as fixed
Call peb_update with:
- id: "{{.PebbleIDPattern}}"
- data: '{"status":"fixed"}'
```
