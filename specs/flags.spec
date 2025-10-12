# Go Command-Line Flag Best Practices and Conventions

## Sources
- Go standard library flag package: https://pkg.go.dev/flag
- Google Go Style Guide: https://google.github.io/styleguide/go/decisions.html
- Community discussions and real-world examples (kubectl, go tool, docker)

## Naming Conventions

### Flag Names (External Interface)
1. **Prefer underscores** over hyphens for multi-word flags
   - Good: `--poll_interval`, `--camera_make`, `--iso_min`
   - Acceptable: `--poll-interval` (common in GNU tools, but not Go convention)
   - Bad: `--pollInterval` (camelCase in flags)

2. **Use descriptive, full words** when possible
   - Go philosophy: "flags have proliferated to point where 'long' names are common case"
   - Good: `--database`, `--verbose`, `--output`
   - Acceptable: `--db` (common abbreviation), `-v` (universally understood shorthand)
   - Consider: Provide both long and short versions for frequently-used flags

3. **Boolean flags should be positive statements**
   - Good: `--verbose`, `--open`, `--facets`
   - Avoid: `--no-verbose`, `--disable-cache` (prefer `--cache` with default true)

4. **Use consistent terminology**
   - If using `--output` in one command, don't use `--out` in another
   - Maintain same flag names across subcommands when they mean the same thing

### Variable Names (Internal Code)
1. **Use camelCase** for Go variables that hold flag values
   ```go
   var (
       pollInterval = flag.Duration("poll_interval", time.Minute, "...")
       cameraMake = flag.String("camera_make", "", "...")
   )
   ```

2. **Place flag variables in their own `var` group** after imports
   ```go
   import (...)

   var (
       // Flag definitions
       dbPath = flag.String("db", "photos.db", "...")
       workers = flag.Int("workers", 4, "...")
   )
   ```

## Syntax and Usage

### Flag Syntax Support (Go flag package)
- Single dash: `-flag`, `-flag=value`, `-flag value`
- Double dash: `--flag`, `--flag=value`, `--flag value`
- Both are **equivalent** in Go (unlike GNU conventions)

### Special Cases
- Boolean flags: `-flag=false` (explicit value required for false)
- Standalone `-` is a non-flag argument
- `--` terminates flag parsing
- `-flag value` syntax not allowed for booleans (ambiguity with shell wildcards)

### Multiple Short Flags
- **Not supported**: Cannot use `-abc` for `-a -b -c`
- This is intentional Go design (use pflag package if needed)

## Documentation Best Practices

### Usage Messages
1. **Be concise but descriptive**
   - Good: "Path to database file"
   - Good: "Number of worker threads (default 4)"
   - Avoid: "db" or "database" (too terse)

2. **Include units and constraints**
   - Good: "Thumbnail size (64, 256, 512, or 1024)"
   - Good: "Maximum aperture (e.g., 5.6)"
   - Good: "Filter by year (e.g., 2025)"

3. **Indicate when flags are required**
   - Good: "Path to database file (required)"
   - Note: flag package doesn't enforce required flags; validate manually

4. **Show common examples**
   - Good: "Camera make (e.g., Canon)"
   - Good: "Time of day (morning, afternoon, evening, night)"

### Help Output
1. **Group related flags** in documentation
   - Temporal filters: year, month, day
   - Equipment filters: camera, lens
   - Technical filters: ISO, aperture, focal length

2. **Provide examples** showing real usage
   ```
   Examples:
     olsen query --db photos.db --year 2025 --color blue
     olsen index ~/Pictures --db photos.db -w 8
   ```

3. **Command-specific help** should be accessible
   - `command --help` shows detailed flags for that command
   - Main help shows summary with key flags

## Subcommand Pattern

### FlagSet Usage
1. **Create separate FlagSet for each subcommand**
   ```go
   func queryCommand(args []string) error {
       fs := flag.NewFlagSet("query", flag.ExitOnError)
       dbPath := fs.String("db", "", "Path to database file (required)")
       // ... more flags
       fs.Parse(args)
   }
   ```

2. **Consistent flag names across subcommands**
   - If multiple commands need `--db`, they should all use `--db`
   - Avoid: `--database` in one command, `--db` in another

3. **Subcommand-specific flags are fine**
   - `query --color blue` (query-specific)
   - `index -w 8` (index-specific)
   - `thumbnail -s 256` (thumbnail-specific)

## Validation and Defaults

### Default Values
1. **Provide sensible defaults** when possible
   - Good: `--db photos.db` (common default)
   - Good: `--workers 4` (reasonable parallelism)
   - Good: `--addr localhost:8080` (common development port)

2. **Empty string for required flags**
   - Use `""` as default, then validate after parsing
   ```go
   dbPath := fs.String("db", "", "Path to database file (required)")
   fs.Parse(args)
   if *dbPath == "" {
       return fmt.Errorf("--db flag is required")
   }
   ```

3. **Zero values that can be validated**
   - Use `0` for required integers, validate `> 0` after parsing
   - Use `nil` pointers for optional values

### Validation Best Practices
1. **Validate immediately after parsing**
   ```go
   fs.Parse(args)
   if *dbPath == "" {
       return fmt.Errorf("--db flag is required")
   }
   if *size != "64" && *size != "256" && *size != "512" {
       return fmt.Errorf("invalid size: %s", *size)
   }
   ```

2. **Provide helpful error messages**
   - Good: "invalid size: 128 (must be 64, 256, 512, or 1024)"
   - Bad: "invalid size"

3. **Check file/directory existence** where appropriate
   ```go
   if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
       return fmt.Errorf("database file not found: %s", *dbPath)
   }
   ```

## Real-World Examples

### Go Standard Tools
```bash
go build -o output -v ./...
go test -v -race -cover ./...
go tool pprof -http=:8080 cpu.prof
```
- Mix of short (`-v`, `-o`) and long (`-race`, `-cover`, `-http`) flags
- Underscores in compound words rare (mostly single words)

### Kubernetes kubectl
```bash
kubectl get pods --namespace default
kubectl logs pod-name --follow
kubectl apply -f config.yaml
```
- Double dashes for long flags
- Short versions for common flags (`-f`, `-n`)
- Consistent naming across subcommands

### Docker
```bash
docker run -d -p 8080:80 --name myapp nginx
docker build -t myimage:latest .
```
- Single dash for short flags
- Double dash for long flags
- Mix of approaches (not Go-based tool)

## Anti-Patterns to Avoid

1. **Don't use flags in library code**
   - Flags should only be in `package main`
   - Libraries should use function parameters or config structs

2. **Don't mix naming conventions**
   - Don't use both `--camera-make` and `--lens_model`
   - Pick one style (underscores preferred for Go)

3. **Don't overload flags**
   - Avoid: `--format json,pretty,color`
   - Better: `--format json --pretty --color`

4. **Don't make boolean flag names confusing**
   - Avoid: `--no-cache`, `--disable-verbose`
   - Better: `--cache` (with sensible default), `--verbose` or `--quiet`

5. **Don't skip validation**
   - The flag package won't validate required flags
   - The flag package won't validate enum values
   - Always validate after `Parse()`

## Summary Checklist

For any Go CLI tool:
- [ ] Flag names use underscores for multi-word names
- [ ] Variable names use camelCase
- [ ] All flags have clear, descriptive usage messages
- [ ] Default values are sensible or explicitly validated
- [ ] Required flags are validated after parsing
- [ ] Enum-like flags list valid values in usage
- [ ] Each subcommand has its own FlagSet
- [ ] Consistent flag names across subcommands
- [ ] Main help shows available commands with key flags
- [ ] Subcommand help available via `command --help`
- [ ] Examples demonstrate real-world usage
- [ ] File/path flags are validated for existence
- [ ] Boolean flags are positive statements
- [ ] No flags in library packages
