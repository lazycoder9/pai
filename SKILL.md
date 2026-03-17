---
name: pai
description: "Manages project ideas, features, tasks, and decisions using the pai CLI. Use when asked to plan work, track tasks, log decisions, create features, manage backlog, or when asking about project state like 'where are we?', 'what should we work on?', 'anything critical?', or 'what's next?'."
---

# Using pai — AI-Native Project Management

`pai` is a CLI tool that stores project knowledge in a `.pai/` folder as markdown files with YAML frontmatter. Use it to manage the development lifecycle: ideas → features → tasks → code.

## Core Concepts

- **Entities**: idea, feature, task, decision — each stored as a markdown file
- **Identity model**: entities have a typed `id` (`T-12`) and a human `slug` (`auth-login`)
- **Knowledge graph**: entities link via `--parent`, forming a chain: Idea → Feature → Task
- **Task pipeline**: tasks move through `backlog → active → done`
- **Reference resolution**: `pai get/edit/delete/start/complete` accept either an `id` or a `slug`

## Commands Reference

### Initialize a project

```bash
pai init --name "project-name"
```

### Add entities

```bash
pai add idea <slug> --body "description" [--tags "t1,t2"] [--priority medium]
pai add feature <slug> --parent <idea-id-or-slug> --body "spec"
pai add task <slug> --parent <feature-id-or-slug> --body "implementation details" [--priority high]
pai add decision <title> --body "context and reasoning"
```

Typed IDs are auto-generated per entity type, starting at `I-1`, `F-1`, `T-1`, `D-1`.

Body can also be provided via stdin for multiline content.

Preferred pattern for multiline bodies:

```bash
pai add task <slug> --parent <feature-id-or-slug> <<'EOF'
Goal: ...

- bullet 1
- bullet 2
EOF
```

Avoid using escaped `\n` inside `--body` strings for long content, because those may be saved literally instead of as real newlines.

### List entities

```bash
pai list ideas [--status raw] [--tag backend]
pai list features
pai list tasks [--status backlog]
pai list decisions
```

### Get entity details

```bash
pai get <ref>                    # auto-detect type
pai get task <ref>               # specific type
pai get task <ref> --all         # include parent chain + children
```

### Edit entities

```bash
pai edit idea <ref> --status explored
pai edit task <ref> --status active      # moves task directory
pai edit feature <ref> --parent <ref> --body "updated spec"
pai edit task <ref> --tags "api,backend" --priority high
pai edit task <ref> --slug auth-login-v2
```

`pai edit` also accepts body content from stdin, which is the preferred way to update long multiline specs:

```bash
pai edit task <ref> <<'EOF'
Goal: ...

Acceptance Criteria:
- item 1
- item 2
EOF
```

Prefer heredoc/stdin over `echo` for multiline text, since `echo` can be inconsistent with escaping and trailing newlines across shells.

### Delete entities

```bash
pai delete idea <ref>
pai delete feature <ref>
pai delete task <ref>
pai delete decision <ref>
```

### Task lifecycle shortcuts

```bash
pai start <task-ref>       # backlog → active
pai complete <task-ref>    # active → done
```

### Project overview

```bash
pai status                  # shows full project tree
```

## Workflows

### Capturing a new idea

1. Run `pai status` to see current state
2. `pai add idea <slug> --body "description"`
3. Refine later with `pai edit idea <ref> --status explored`

### Planning a feature from an idea

1. `pai add feature <slug> --parent <idea-id-or-slug> --body "feature spec"`
2. Decompose into tasks:
   ```bash
   pai add task <slug1> --parent <feature-id-or-slug> --body "step 1" --priority high
   pai add task <slug2> --parent <feature-id-or-slug> --body "step 2"
   ```
3. `pai edit idea <ref> --status tasks_generated`

### Working on tasks

1. `pai list tasks --status backlog` to see what's next
2. `pai start <task-ref>` to begin work
3. Implement the code
4. `pai complete <task-ref>` when done

### Logging a decision

```bash
pai add decision "use-postgres-for-storage" --body "Context: need relational data. Alternatives: SQLite, MongoDB. Chose Postgres for scalability."
```

## Answering Project Questions

When the user asks conversational questions about the project, use `pai` to answer them:

| User says | What to do |
|---|---|
| "Where are we?" / "What's the status?" | Run `pai status` and summarize the project tree |
| "What should we work on?" / "What's next?" | Run `pai list tasks --status backlog`, then highlight by priority |
| "Anything critical?" / "Any blockers?" | Run `pai list tasks --status active` and `pai list tasks --status backlog`, look for `--priority high` items |
| "What ideas do we have?" | Run `pai list ideas` and summarize |
| "Tell me about X" | Run `pai get <ref> --all` to show full context chain |
| "How's feature X going?" | Run `pai get feature <ref> --all` to show the feature and its child tasks |

Always interpret these questions through `pai` data first, then provide a natural-language summary of the findings.

## Best Practices

- Always check `pai status` before starting work to understand project state
- Link entities with `--parent` to maintain traceability
- Prefer typed IDs in branches and commits, but keep slugs short and descriptive
- Use `pai get <ref> --all` to understand full context before implementing
- Log architectural decisions immediately so they aren't lost
- Keep slugs short and descriptive: `user-auth`, `api-rate-limiting`
