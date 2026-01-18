# AGENTS.md - Code Style & Development Guide

## Build & Test

The project uses a Makefile for common development tasks:

- `make` - Build the binary to `bin/peb` (default target)
- `make test` - Run all tests with verbose output
- `make test-coverage` - Run tests with coverage report (generates
  coverage.html)
- `make clean` - Remove build artifacts and coverage files
- `make install` - Install the binary to GOPATH/bin
- `make fmt` - Format code with go fmt
- `make vet` - Run go vet for static analysis
- `make mod-tidy` - Tidy go.mod dependencies
- `make deps` - Download dependencies
- `make help` - Show all available targets

## Running the Binary

After building with `make`, the binary is available at `bin/peb`:

```bash
# Run directly
./bin/peb <command>

# Or install to PATH for easier access
make install
# Now you can use: peb <command>
```

## Module Structure

`pebbles` is a task tracking CLI tool that manages "pebs" (tasks/bugs/features)
as markdown files with YAML frontmatter.

- `cmd/peb/` - CLI app entry point (urfave/cli)
- `internal/commands/` - Command implementations (init, new, read, update,
  query, prime)
- `internal/config/` - Configuration loading from `.pebbles/config.toml`
- `internal/peb/` - Core data model, validation, file I/O (YAML frontmatter)
- `internal/store/` - Storage layer, ID generation, lookups

## Go Code Style

### Structs & Interfaces

- Use dependency injection via concrete types unless you have multiple
  implementations
- Don't create interfaces unless you have multiple implementations or need to
  mock for tests
- Keep structs simple, focus on single responsibility
- Use pointer receivers for methods that modify state
- Use value receivers for methods that don't modify state
