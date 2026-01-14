# wk

A CLI tool to simplify git worktree management with automatic file copying and post-creation hooks.

## Installation

### Quick install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/lucas-stellet/wk/main/install.sh | sh
```

### Homebrew (macOS/Linux)

```bash
brew tap lucas-stellet/wk
brew install wk
```

### Go install

```bash
go install github.com/lucas-stellet/wk@latest
```

### Build from source

```bash
git clone https://github.com/lucas-stellet/wk.git
cd wk
go build -o wk .
```

## Usage

### Initialize configuration

```bash
wk init
```

This creates a `.wk.yaml` file interactively, asking for:
- Files/directories to copy to new worktrees
- Post-creation hooks to run

### Create a new worktree

```bash
wk new feature-branch
```

This will:
1. Run `git worktree add feature-branch`
2. Copy files listed in `.wk.yaml`
3. Execute post-creation hooks

### List worktrees

```bash
wk list
# or
wk ls
```

### Remove a worktree

```bash
wk remove feature-branch
# or
wk rm feature-branch
```

## Requirements

- Must be run inside a git repository
- If `.wk.yaml` is not found, commands will show a hint suggesting `wk init` but will continue execution
- If `.wk.yaml` contains invalid YAML, commands will fail with an error

## Configuration

Create a `.wk.yaml` in your project root:

```yaml
# Files and directories to copy from source to new worktree
copy:
  - .env
  - .env.local
  - tmp/

# Commands to run after creating the worktree (in the new worktree directory)
post_hooks:
  - npm install
  - cp .env.example .env
```

## Example workflow

```bash
# In your main project directory
cd my-project

# Create configuration
wk init

# Create a new worktree for a feature
wk new my-feature

# Work on the feature...
cd ../my-feature

# When done, remove the worktree
cd ../my-project
wk rm my-feature
```

## License

MIT
