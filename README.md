# git-herd 🐑

A decent, not bad, concurrent Git repository management tool written in Go. git-herd allows you to perform bulk `fetch` or `pull` operations across multiple Git repositories in a directory tree.

Because I'm lazy and because any given time I have more than 300 git repos locally I needed a fast way to fetch/pull changes in bulk.

## Features

- 🚀 **Concurrent Processing**: Process multiple repositories in parallel with configurable worker pools
- 🎯 **Smart Repository Discovery**: Automatically finds all Git repositories in directory trees
- 🛡️ **Safety First**: Skip dirty repositories to prevent conflicts
- 🔍 **Dry Run Mode**: Preview operations before execution
- ⚡ **Fast**: Built with Go for maximum performance
- 📊 **Detailed Reporting**: Clear progress and result reporting
- ⚙️ **Configurable**: Extensive configuration options via flags or config file
- 🚨 **Graceful Shutdown**: Handles interrupts cleanly
- 📝 **Structured Logging**: Built-in logging with configurable verbosity

## Installation

### From Source

```bash
git clone https://github.com/entro314-labs/git-herd
cd git-herd
make build
sudo make install
```

### Using Go Install

```bash
go install github.com/entro314-labs/git-herd/cmd/git-herd@latest
```

## Usage

### Basic Examples

```bash
# Fetch all repositories in current directory
git-herd

# Pull all repositories in a specific directory
git-herd -o pull ~/Projects

# Dry run to see what would happen
git-herd -n -o pull ~/Projects

# Use more workers for faster processing
git-herd -w 10 ~/Projects

# Verbose output for debugging
git-herd -v ~/Projects
```

### Command Line Options

```
Usage:
  git-herd [path] [flags]

Flags:
  -e, --exclude strings     Directories to exclude (default [.git,node_modules,vendor])
  -n, --dry-run            Show what would be done without executing
  -h, --help               help for git-herd
  -o, --operation string   Operation to perform: fetch or pull (default "fetch")
  -r, --recursive          Process repositories recursively (default true)
  -s, --skip-dirty         Skip repositories with uncommitted changes (default true)
  -t, --timeout duration   Overall operation timeout (default 5m0s)
  -v, --verbose            Enable verbose logging
  -w, --workers int        Number of concurrent workers (default 5)
```

### Configuration File

Create a `git-herd.yaml` file in your working directory or `~/.config/git-herd/`:

```yaml
operation: fetch
workers: 10
skip-dirty: true
recursive: true
verbose: false
timeout: 10m
exclude:
  - .git
  - node_modules
  - vendor
  - target
  - dist
```

## Operations

### Fetch vs Pull

- **Fetch** (`-o fetch`): Downloads changes from remote without merging (safe, default)
- **Pull** (`-o pull`): Downloads and merges changes (requires clean working directory)

### Safety Features

- **Dirty Repository Handling**: By default, repositories with uncommitted changes are skipped when pulling
- **Timeout Protection**: Configurable timeout prevents hanging operations
- **Graceful Shutdown**: SIGINT/SIGTERM handling allows clean cancellation
- **Error Isolation**: Failures in one repository don't affect others

## Output Format

git-herd provides clear, structured output:

```
📊 Processing Results:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ project1 (/path/to/project1) [main@origin] - 245ms
✅ project2 (/path/to/project2) [develop@origin] - 180ms
❌ project3 (/path/to/project3): repository has uncommitted changes (skipped)
✅ project4 (/path/to/project4) [main@origin] - 320ms
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📈 Summary: 3 successful, 1 failed, 4 total
```

## TUI Mode

By default, git-herd runs with a beautiful Terminal User Interface (TUI) that shows:
- Real-time progress with a progress bar
- Live status updates for each repository
- Colored output for success/failure states
- Summary statistics

To disable TUI and use plain text output:
```bash
git-herd --plain ~/Projects
```

### Report Generation

Generate detailed reports of operations:
```bash
# Save a detailed report to file
git-herd --save-report report.txt ~/Projects

# Show full summary of all repositories
git-herd --full-summary ~/Projects
```

## Advanced Usage

### Working with Large Repository Collections

For better performance with many repositories:

```bash
# Increase workers and timeout for large collections
git-herd -w 20 -t 15m ~/all-projects

# Process only direct subdirectories (not recursive)
git-herd -r=false ~/Projects
```

### Excluding Specific Directories

```bash
# Exclude build artifacts and dependencies
git-herd -e node_modules,target,dist,vendor ~/Projects

# Use with specific operations
git-herd -o pull -e ".git,tmp,cache" ~/Projects
```

### Integration with Shell

Add to your shell profile for quick access:

```bash
# Fetch all repos in current directory
alias gf='git-herd'

# Pull all repos in current directory
alias gp='git-herd -o pull'

# Fetch all repos in Projects directory
alias gfp='git-herd ~/Projects'
```

## Performance

- **Concurrent**: Processes multiple repositories simultaneously
- **Efficient**: Pure Go implementation with minimal dependencies
- **Scalable**: Handles hundreds of repositories efficiently
- **Resource-Aware**: Configurable worker pools prevent resource exhaustion

## Error Handling

git-herd provides detailed error reporting and handles common scenarios:

- **Network timeouts**: Configurable timeout handling
- **Authentication failures**: Clear error messages for auth issues
- **Dirty repositories**: Safe skipping with clear reporting
- **Missing remotes**: Graceful handling of repositories without remotes
- **Permission issues**: Clear error reporting for access problems

## Building from Source

Requirements:
- Go 1.25 or later

```bash
git clone https://github.com/entro314-labs/git-herd
cd git-herd
make deps
make build
```

### Cross-platform Builds

```bash
# Build for all platforms
make build-all

# Or build for specific platforms manually:
# make build-darwin-amd64
# make build-linux-amd64
# make build-windows-amd64
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [go-git](https://github.com/go-git/go-git) for Git operations
- CLI powered by [Cobra](https://github.com/spf13/cobra)
- Configuration management via [Viper](https://github.com/spf13/viper)
