# Using peb (Pebbles Task Tracker)

You are working with "peb" (Pebbles), an agent-first tool for tracking tasks, bugs, features, and epics.

## ⚠️ CRITICAL REQUIREMENT

**ALL NON-TRIVIAL WORK MUST BE TRACKED AS A PEB**

Before doing ANY non-trivial task, bug fix, feature, or code change:
1. Create a peb to track it (for complex work, create multiple subtasks followed by an overall epic blocked by all the subtasks)
2. Update its status as you work
3. Mark it as fixed when complete

Trivial work (simple, single-step tasks that can be completed in < 3 minutes) can be done without a peb.

**Rules:**

1. Every non-trivial task, bug, feature, or code change must be tracked as a peb
2. Create a peb before starting the work
3. Update peb status throughout the lifecycle
4. Do not mark pebs as `fixed` until all dependencies (`blocked-by`) are also `fixed`
5. Use `blocked-by` to establish clear dependencies between related work
6. For complex work, break it down into smaller task pebs and create an `epic` peb blocked by all the task pebs (epic remains `in-progress` until all tasks are `fixed`)

## Core Concepts

**Peb Structure:** A "peb" represents a task/bug/feature/epic with these fields:

- `id`: Unique identifier (format: `{{.PebbleIDPattern}}` where `{{.PebbleIDSuffix}}` is a random ID)
- `title`: Short description
- `type`: One of: `bug`, `feature`, `epic`, `task`
- `status`: One of: `new`, `in-progress`, `fixed`, `wont-fix`
- `created`/`changed`: timestamps
- `blocked-by`: List of peb IDs that must be fixed before this peb can be marked as fixed (dependencies/subtasks)
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

**Destructive operations:**

- **DO NOT use `peb_delete`** unless the user explicitly asks for it. This command permanently deletes pebs and their data cannot be recovered. Always confirm with the user before using this command.

**Tracking dependencies with blocked-by:**

The `blocked-by` field establishes dependencies between pebs:
- Setting {{.PebbleIDPattern}} as blocked-by {{.PebbleIDPattern2}} means {{.PebbleIDPattern2}} is a prerequisite or subtask of {{.PebbleIDPattern}}
- {{.PebbleIDPattern}} cannot be marked as `fixed` until all pebs in its `blocked-by` list are also `fixed`
- Use `peb_read` to find a peb's dependencies (must be completed before this peb can be marked as fixed)
  - Use `peb_query` with `filters: ["id:(<id>|...)"]` to get all the titles of the dependencies
  - Use `peb_read` to get their full details

## Example: Bug Fix without Dependencies

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

## Example: Creating an Epic

When tracking complex work that requires multiple tasks, create an epic:

1. First, create all the subtask pebs:
   - Call `peb_new` for each subtask with type `task` or `feature`
   - Leave `blocked-by` empty

2. Create the epic peb that tracks the overall goal:
   - Call `peb_new` with type `epic`
   - Set `blocked-by` to list all the subtask peb IDs

## Writing Good Descriptions

**Good Task/Bug/Feature Descriptions:**

- Be specific and actionable - describe what needs to be done
- Include context - why is this task needed?
- Add acceptance criteria - how will you know it's complete?
- Reference relevant files, code locations, or issues
- Include examples or expected behavior where helpful
- Keep it concise but complete - enough info for another agent to execute

**Good Epic Descriptions:**

- Focus on the "what" and "why" - the overall goal
- Break down into clear, testable components
- Link to related tasks using `blocked-by`
- Include success criteria for the epic
