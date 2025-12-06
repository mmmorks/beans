# What we're building

This is going to be a small CLI app that interacts with a .beans/ directory that stores "issues" (like in an issue tracker) as markdown files with front matter. It is meant to be used as part of an AI-first coding workflow.

- This is an agentic-first issue tracker. Issues are called beans.
- Projects can store beans (issues) in a `.beans/` subdirectory.
- The executable built from this project here is called `beans` and interacts with said `.beans/` directory.
- The `beans` command is designed to be used by a coding agent (Claude, OpenCode, etc.) to interact with the project's issues.
- `.beans/` contains markdown files that represent individual beans.
- Every markdown file found is considered a bean, but they can be grouped in subdirectories, nested as the user sees fit. This allows the user to structure their issues however they want (eg. around projects, epics, etc.)
- The individual bean filenames start with a string-based ID (use 3-character NanoID here so things stay mergable), optionally followed by a dash and a short description
  (mostly used to keep things human-editable). Examples for valid names: `f7g.md`, `f7g-user-registration.md`.

# Rules

- ONLY make commits when I explicitly tell you to do so.
- When making commits, provide a meaningful commit message. The description should be a concise bullet point list of changes made.

# Bean structure

- Each bean is a markdown file with front matter.
- The front matter contains metadata about the bean, including:
  - `title`: a human-readable, one-line title for the bean
  - `status`: the current status of the bean (e.g., `open`, `in-progress`, `done`)
  - `created_at`: timestamp of when the bean was created
  - `updated_at`: timestamp of the last update to the bean

# Building

- `mise build` to build a `./beans` executable

# Testing

- Use `go run .` instead of building the executable first.
- All commands support the `--beans-path` flag to specify a custom path to the `.beans/` directory. Use this for testing (instead of spamming the real `.beans/` directory).
