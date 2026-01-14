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

- `cmd/` - Cobra commands (init, new, list, remove). Each file registers one command via `init()`.
- `internal/config/` - Parses `.wk.yaml` configuration; searches upward from current directory.
- `internal/worktree/` - Wraps git worktree operations (add, list, remove) via `exec.Command`.
- `internal/hooks/` - File/directory copying and shell command execution for post-hooks.

### Flow: `wk new <branch>`

1. `cmd/new.go` calls `worktree.Add()` to create worktree via `git worktree add`
2. Loads `.wk.yaml` using `config.FindConfig()` (walks up directory tree)
3. Calls `hooks.CopyFiles()` to copy specified files from source to new worktree
4. Calls `hooks.RunPostHooks()` to execute shell commands in the new worktree directory

### Configuration (.wk.yaml)

```yaml
copy:          # Files/dirs to copy from source worktree
  - .env
post_hooks:    # Shell commands to run in new worktree
  - npm install
```
