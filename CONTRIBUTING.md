# Contributing to Olsen

Thank you for your interest in contributing to Olsen! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Be respectful, constructive, and professional in all interactions.

## Getting Started

### Prerequisites

- Go 1.21 or later
- SQLite 3
- For RAW support: libraw library (`brew install libraw` on macOS)

### Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/adewale/olsen.git
   cd olsen
   ```

2. **Build the project**:
   ```bash
   # Standard build (no RAW support)
   make build

   # With LibRaw support (requires libraw installed)
   make build-raw
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Create a test database**:
   ```bash
   ./bin/olsen index testdata/dng --db test.db --w 2
   ```

## Development Workflow

### Running Tests

The project uses a two-tier testing strategy to avoid CGO/LibRaw dependencies in CI:

```bash
# Quick tests (recommended for development)
make test

# This runs with CGO_ENABLED=0 to test without SQLite dependencies
# Database-dependent tests are skipped
# Command-line parsing and URL tests pass

# Run with coverage (non-database tests)
CGO_ENABLED=0 go test -coverprofile=coverage.out ./cmd/olsen/...
go tool cover -html=coverage.out

# Run specific package tests
CGO_ENABLED=0 go test -v ./cmd/olsen/
CGO_ENABLED=0 go test -v ./internal/query/ -run "URL|Parsing"

# Run benchmarks
go test -bench=. -benchtime=3s ./internal/indexer/
```

**Note**: Integration tests that require SQLite need `CGO_ENABLED=1`. The CI pipeline runs tests with `CGO_ENABLED=0` to avoid requiring database dependencies. Contributors should ensure their changes pass the `make test` target before submitting PRs.

### Code Style

- **Format all code** with `gofmt`:
  ```bash
  gofmt -w .
  gofmt -s -w .  # Apply simplifications
  ```

- **Run go vet**:
  ```bash
  go vet ./...
  ```

- **Follow Go conventions**:
  - Exported names should have doc comments
  - Doc comments start with the name being documented
  - Error strings are lowercase (no ending punctuation)
  - Use receiver names consistently (1-2 letters)

### Project Structure

```
olsen/
├── cmd/olsen/           # CLI entrypoint
├── internal/            # Private packages
│   ├── database/        # SQLite operations
│   ├── indexer/         # Photo indexing engine
│   ├── explorer/        # Web UI server
│   ├── query/           # Query engine & faceted search
│   └── quality/         # Thumbnail QA tools
├── pkg/models/          # Public data structures
├── testdata/            # Test fixtures
├── specs/               # Technical specifications
└── docs/                # Documentation

```

### Key Design Principles

1. **Read-Only Guarantee**: Olsen NEVER modifies photo files
2. **State Machine Navigation**: Facets are independent dimensions, not hierarchical
3. **Database Portability**: Single SQLite file contains all metadata & thumbnails
4. **Concurrent Processing**: Worker pool pattern for indexing
5. **Aspect Ratio Preservation**: Thumbnails constrain longest edge, not forced squares

## Contributing Code

### Pull Request Process

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write clear, focused commits
   - Add tests for new functionality
   - Update documentation as needed
   - Ensure all tests pass

3. **Format and lint**:
   ```bash
   gofmt -w .
   go vet ./...
   go test ./...
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Add feature: description of what you added"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**:
   - Provide a clear description of the changes
   - Reference any related issues
   - Explain the motivation and context

### Commit Message Guidelines

- Use the imperative mood ("Add feature" not "Added feature")
- Keep the first line under 72 characters
- Provide context in the commit body if needed

**Examples**:
```
Add color classification for B&W photos

- Implement saturation-first logic to detect grayscale images
- Add tests for B&W detection edge cases
- Update specs/dominant_colours.spec with new algorithm

Closes #42
```

### Testing Guidelines

- **Write tests for all new functionality**
- **Follow table-driven test patterns** where appropriate
- **Use descriptive test names**: `TestYearFacetPreservesMonthFilter`
- **Test edge cases** and error conditions
- **Aim for >70% code coverage**

**Example test structure**:
```go
func TestColorClassification(t *testing.T) {
    tests := []struct {
        name     string
        input    Color
        expected string
    }{
        {"pure red", Color{R: 255, G: 0, B: 0}, "red"},
        {"grayscale", Color{R: 128, G: 128, B: 128}, "gray"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ClassifyColor(tt.input)
            if result != tt.expected {
                t.Errorf("got %s, want %s", result, tt.expected)
            }
        })
    }
}
```

## Documentation

### Adding Package Documentation

Every package must have a doc comment:

```go
// Package indexer implements the core photo indexing engine with concurrent processing.
//
// It extracts EXIF metadata, generates aspect-ratio-preserving thumbnails, analyzes
// color palettes, computes perceptual hashes, and infers additional metadata.
package indexer
```

### Adding Function Documentation

Exported functions must have doc comments:

```go
// ExtractMetadata extracts EXIF metadata from DNG and JPEG files.
// It returns basic file metadata if EXIF extraction fails.
func ExtractMetadata(filePath string) (*PhotoMetadata, error) {
    // ...
}
```

### Updating Specifications

If your change affects the design or behavior:

1. Update relevant spec files in `specs/`
2. Update `docs/LESSONS_LEARNED.md` if you discovered something valuable
3. Add architectural decision records to `docs/` for significant changes

## Reporting Issues

### Bug Reports

Include:
- **Olsen version** (`./bin/olsen version`)
- **Operating system and version**
- **Steps to reproduce**
- **Expected vs actual behavior**
- **Sample files** (if relevant and safe to share)
- **Error messages and logs**

### Feature Requests

Describe:
- **Use case**: What problem does this solve?
- **Proposed solution**: How should it work?
- **Alternatives considered**: Other approaches you've thought of
- **Additional context**: Screenshots, examples, etc.

## Questions?

- Check the [README](README.md) for basic usage
- Review [CLAUDE.md](CLAUDE.md) for developer guidance
- Browse specs in `specs/` for technical details
- Check existing issues for similar questions

## License

By contributing to Olsen, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Your contributions make Olsen better for everyone. We appreciate your time and effort!
