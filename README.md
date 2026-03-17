<p align="center">
  <img src="./logo.svg" alt="pai — project ainager">
</p>

# pai

`pai` = `Project AInager`.

Haha, did you get it? No? Never mind.

`pai` is a small CLI for keeping project memory in plain files. It stores ideas, features, tasks, and decisions inside a `.pai/` folder so the important stuff does not disappear into chat history.

I made it for myself because this is how I like to work:

- chat for thinking
- files for memory
- one tiny CLI to keep both connected

It is opinionated, but not in a "you are using software wrong" way. More in a "this workflow makes my brain quieter" way. If it helps you too, nice.

## What it does

`pai` creates a `.pai/` folder in your project:

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

Each item is just Markdown with a bit of metadata. The usual flow is:

```text
idea -> feature -> task
```

That means you can inspect it, edit it, diff it, and commit it like normal project files instead of hiding everything inside some tool.

Each entity now has:

- a typed ID like `T-12` or `I-2`
- a human slug like `auth-login`

`pai` shows both, stores both, and lets you look items up by either one.

## Why it exists

AI-assisted work has a memory problem.

Chat is good for exploration, but bad at being the source of truth. Context gets buried, decisions get rediscovered, and tasks lose the reason they exist.

`pai` fixes that by giving the project a real memory on disk.

It also makes working with AI agents easier. If your agent has access to `pai`, it can check status, create tasks, log decisions, and answer project questions without you manually translating every thought into files. That is what the included [`SKILL.md`](./SKILL.md) is for.

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

Add an idea, turn it into a feature, then break it into a task:

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

See the project tree:

```bash
pai status
```

Example output:

```text
💡 I-1 inbox-zero-for-agents raw
└── 🔧 F-1 lightweight-project-memory spec
    └── 📌 T-1 scaffold-pai-folder backlog
```

Inspect an item:

```bash
pai get task T-1
```

Move the task forward:

```bash
pai start T-1
pai complete T-1
```

## Commands

- `pai init` creates a `.pai/` workspace
- `pai add idea|feature|task|decision <slug>` creates an item with an auto-generated typed ID
- `pai edit ...` updates metadata or content
- `pai delete ...` removes an item
- `pai list ideas|features|tasks|decisions` lists items with filters
- `pai get <ref>` or `pai get <type> <ref>` shows item details by ID or slug
- `pai status` prints the project tree
- `pai start <task-ref>` moves a task to `active`
- `pai complete <task-ref>` moves a task to `done`

## Notes

- local-first
- plain files
- git-friendly
- small command surface
- useful for humans, easy for agents

This is intentionally small. It is not trying to become a giant project-management platform. It is just a sharp little tool for people who want structured project memory without a lot of ceremony.
