# Using peb (Pebbles Task Tracker)

You are working with "peb" (Pebbles), an agent-first Go CLI tool for tracking
tasks, bugs, features, and epics.

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

- Pebs with status `new` or `in-progress` are "open". Query with `peb query status:open`.
- Pebs with status `fixed` or `wont-fix` are "closed". Query with `peb query status:closed`.

## Common Workflows

### Create a new peb

```bash
echo '{"title":"Fix login bug","content":"Users cannot log in","type":"bug"}' | peb new
```

Required: `title`, `content` Optional: `type` (default: `bug`), `blocked-by`
(array of peb IDs)

### Create a dependent peb (blocked by another)

```bash
echo '{"title":"Fix UI","content":"...","blocked-by":["{{.PebbleIDPattern}}"]}' | peb new
```

### Read a peb

```bash
peb read {{.PebbleIDPattern}}
```

Returns full peb data as JSON.

### Update a peb

```bash
peb update {{.PebbleIDPattern}} '{"status":"in-progress"}'
peb update {{.PebbleIDPattern}} '{"title":"New title"}'
peb update {{.PebbleIDPattern}} '{"blocked-by":["{{.PebbleIDPattern2}}","{{.PebbleIDPattern3}}"]}'
```

### Query pebs

```bash
# List all pebs
peb query

# Filter by status
peb query status:new

# Filter by open status (new OR in-progress)
peb query status:open

# Filter by closed status (fixed OR wont-fix)
peb query status:closed

# Filter by type
peb query type:feature

# Find pebs blocked by a specific peb
peb query blocked-by:{{.PebbleIDPattern}}

# Combine filters (implicit AND)
peb query status:new type:bug

# Output specific fields only
peb query --fields id,title
```

### Cleanup pebs

```bash
peb cleanup
```

**⚠️ CRITICAL WARNING: THIS COMMAND DELETES DATA**

**DO NOT run `peb cleanup` unless the user explicitly asks for it.** This command permanently deletes pebs and their data cannot be recovered.

This command removes all closed pebs.

Always confirm with the user before running this command.

## Best Practices

**Before starting work:**

1. Use `peb query status:new` to find work
2. Read the full peb with `peb read <id>` to understand requirements

**While working:**

1. Mark pebs as `in-progress` when starting work
2. Update the peb's content if requirements change
3. Mark as `fixed` when completed

**Tracking dependencies:**

- Use `blocked-by` when work depends on another peb
- Use `peb read <id>` to find a peb's dependencies (must be completed before
  this peb can be marked as fixed)
- Query with `blocked-by:<id>` to find pebs blocked by a specific task

## Example Workflow

```bash
# 1. Find a bug to work on
peb query status:open type:bug

# 2. Read details
peb read {{.PebbleIDPattern}}

# 3. Mark as in-progress
peb update {{.PebbleIDPattern}} '{"status":"in-progress"}'

# 4. Do the work...

# 5. Mark as fixed
peb update {{.PebbleIDPattern}} '{"status":"fixed"}'
```
