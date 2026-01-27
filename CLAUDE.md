# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build the binary
go build -o wk .

# Install globally
go install .

# Run directly
go run . <command>

# Format code
go fmt ./...

# Run linter (if using)
go vet ./...
```

Note: This project has no tests currently.

## Architecture

`wk` is a CLI tool that wraps `git worktree` commands with automatic file copying and post-creation hooks. Built with Cobra for CLI handling.

### Package Structure

- `cmd/` - Cobra commands (init, new, list, remove, switch). Each file registers one command via `init()`.
- `internal/config/` - Parses `.wk.yaml` configuration; searches upward from current directory.
- `internal/worktree/` - Wraps git worktree operations (add, list, remove) and branch listing via `exec.Command`.
- `internal/hooks/` - File/directory copying and shell command execution for post-hooks.
- `internal/selector/` - Interactive fuzzy selection UI using bubbletea/bubbles for branch and worktree selection.

### Flow: `wk new [branch]`

1. If no branch specified, opens interactive selector (`selector.SelectOrCreate()`) using bubbletea
2. `cmd/new.go` calls `worktree.Add()` to create worktree via `git worktree add`
3. Loads `.wk.yaml` using `config.FindConfig()` (walks up directory tree)
4. Calls `hooks.CopyFiles()` to copy specified files from source to new worktree
5. Calls `hooks.RunPostHooks()` to execute shell commands in the new worktree directory

### Interactive Selection

Commands `new`, `switch`, and `remove` support interactive mode when called without arguments:
- Uses bubbletea/bubbles for TUI with fuzzy filtering
- `selector.SelectOrCreate()` - for branch selection with option to create new
- `selector.SelectWorktree()` - for worktree selection

### Configuration (.wk.yaml)

```yaml
copy:          # Files/dirs to copy from source worktree
  - .env
post_hooks:    # Shell commands to run in new worktree
  - npm install
```
