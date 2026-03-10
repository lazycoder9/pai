# pai

`pai` is a tiny CLI for running projects out of plain files instead of stuffing everything into chat history, sticky notes, or whatever PM tool you are currently pretending to enjoy.

I built it for myself because this is how I like to work:

- chat for thinking
- files for memory
- one small CLI to keep ideas, features, tasks, and decisions connected

It is opinionated on purpose. If that clicks for you, use it. If it does not, use something else. No hard feelings, but this tool is not going to apologize for having a personality.

It also exists to make working with AI agents less annoying. The idea is simple: give the agent a real project memory and a tiny CLI, then add a `pai` skill so it can manage project state without you micromanaging every step.

## What it does

`pai` creates and manages a `.pai/` folder in your project:

```text
.pai/
  context.md
  architecture.md
  roadmap.md
  ideas/
  features/
  tasks/
    backlog/
    active/
    done/
  decisions/
  state.json
```

Inside that folder, everything is plain Markdown plus a little metadata. The model is simple:

```text
idea -> feature -> task
```

That gives you a lightweight project memory you can inspect, edit, diff, and commit like normal files.

## Why this exists

Most AI-assisted work has a memory problem.

The chat is good at exploration, but terrible at being the source of truth. Important context gets buried, decisions get rediscovered, and tasks lose the reason they exist.

`pai` is my fix for that. Keep the thinking in chat. Keep the durable project state in files.

Another reason: I wanted agents to be able to operate on project state directly. If your agent can use `pai`, it can check status, create features, move tasks, and log decisions without waiting for you to manually translate everything into files.

## Use with agents

`pai` was built specifically to make work with AI agents more convenient.

Instead of keeping project state trapped in chat, you can give the agent a skill that uses this CLI and let it:

- inspect project status
- create ideas, features, tasks, and decisions
- move tasks through backlog, active, and done
- answer "where are we?" style questions from actual project data

This repo now includes a [`SKILL.md`](./SKILL.md) you can drop into an agent setup so the tool can be used without your constant involvement.

## Install

```bash
go install github.com/lazycoder9/pai@latest
```

Or build it locally:

```bash
git clone https://github.com/lazycoder9/pai.git
cd pai
go build -o pai .
```

## Quick start

Initialize a project:

```bash
pai init --name "demo"
```

Add an idea, turn it into a feature, then break it into tasks:

```bash
pai add idea inbox-zero-for-agents \
  --tags cli,agents \
  --body "A place to dump work without losing context."

pai add feature lightweight-project-memory \
  --parent inbox-zero-for-agents \
  --body "Turn the idea into a structured workflow."

pai add task scaffold-pai-folder \
  --parent lightweight-project-memory \
  --priority high \
  --body "Create the folder layout and seed files."
```

See the current tree:

```bash
pai status
```

Example output:

```text
💡 inbox-zero-for-agents raw
└── 🔧 lightweight-project-memory spec
    └── 📌 scaffold-pai-folder backlog
```

Inspect a specific item:

```bash
pai get task scaffold-pai-folder
```

Move work forward:

```bash
pai start scaffold-pai-folder
pai complete scaffold-pai-folder
```

## Commands

- `pai init` initializes a `.pai/` workspace
- `pai add idea|feature|task|decision <slug>` creates a new item
- `pai edit ...` updates metadata or body content
- `pai delete ...` removes an item
- `pai list ideas|features|tasks|decisions` lists items with optional filters
- `pai get <slug>` or `pai get <type> <slug>` shows a single item
- `pai status` prints the project tree
- `pai start <task>` moves a task to `active`
- `pai complete <task>` moves a task to `done`

## Design constraints

- local-first
- plain files
- git-friendly
- small command surface
- useful for humans, easy for agents

## Current state

This is early, but usable. The point is not to be a giant project-management platform. The point is to be a sharp little tool for people who want structured project memory without ceremony.

If that sounds good, steal it.
