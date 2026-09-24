# Package Size, Program Structure, and CLIs

> Sources: source/google-go-styleguide/best-practices.md (Package size); source/uber-go-style/style.md (Avoid init(), Exit in Main); https://pkg.go.dev/flag
> Authority: advisory
> Last verified: 2026-09-10

Detailed guidance on package splitting, avoiding init(), the run() pattern, and
CLI structure.

## Contents

- [When to Split a Package](#when-to-split-a-package)
- [Avoiding init()](#avoiding-init)
- [Exit in Main](#exit-in-main)
- [Command-Line Interfaces](#command-line-interfaces)

## When to Split a Package

```
Is the package getting too large?
├─ Can you describe its purpose in one sentence?
│  ├─ No → Split by responsibility
│  └─ Yes → Keep it, but check below
├─ Do files in the package never import each other's unexported symbols?
│  └─ Yes → Those files could be separate packages
├─ Does the package have distinct user groups using different parts?
│  └─ Yes → Split along user boundaries
└─ Is the godoc page overwhelming?
   └─ Yes → Split to improve discoverability
```

### When NOT to Split

- Don't split just because a file is long — large files in a focused package are
  fine
- Don't create packages with only one type or function
- Don't split if it would create circular dependencies
- Avoid splitting internal helpers into a `util` or `internal/helpers` package

### When to Combine Packages

- If client code likely needs two types to interact, keep them together
- If types have tightly coupled implementations
- If users would need to import both packages to use either meaningfully

### File Organization

No "one type, one file" convention in Go. Files should be focused enough to know
which file contains something and small enough to find things easily.

---

## Avoiding init()

Prefer explicit functions over `init()`:

```go
// Bad: init() with I/O and environment dependencies
var _config Config

func init() {
    cwd, _ := os.Getwd()
    raw, _ := os.ReadFile(path.Join(cwd, "config.yaml"))
    yaml.Unmarshal(raw, &_config)
}
```

```go
// Good: Explicit function for loading config
func loadConfig(name string) (Config, error) {
    raw, err := os.ReadFile(name)
    if err != nil {
        return Config{}, err
    }
    var config Config
    if err := yaml.Unmarshal(raw, &config); err != nil {
        return Config{}, fmt.Errorf("parse %s: %w", name, err)
    }
    return config, nil
}
```

**Acceptable uses of init():**
- Complex expressions that cannot be single assignments
- Pluggable hooks (e.g., `database/sql` dialects, encoding registries)
- Deterministic precomputation

---

## Exit in Main

Call `os.Exit` or `log.Fatal*` **only in `main()`**. All other functions should
return errors to signal failure.

**Why this matters:**
- Non-obvious control flow: Any function can exit the program
- Difficult to test: Functions that exit also exit the test
- Skipped cleanup: `defer` statements are skipped

```go
// Bad: log.Fatal in helper function
func readFile(path string) string {
    f, err := os.Open(path)
    if err != nil {
        log.Fatal(err)  // Exits program, skips defers
    }
    b, err := io.ReadAll(f)
    if err != nil {
        log.Fatal(err)
    }
    return string(b)
}
```

```go
// Good: Return errors, let main() decide to exit
func main() {
    body, err := os.ReadFile(path)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(body))
}
```

### The run() Pattern

Prefer to call `os.Exit` or `log.Fatal` **at most once** in `main()`. Extract
business logic into a separate function that returns errors.

```go
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}

func run() error {
    args := os.Args[1:]
    if len(args) != 1 {
        return errors.New("missing file")
    }

    f, err := os.Open(args[0])
    if err != nil {
        return err
    }
    defer f.Close()  // Will always run

    b, err := io.ReadAll(f)
    if err != nil {
        return err
    }

    // Process b...
    return nil
}
```

**Benefits of the `run()` pattern:**
- Short `main()` function with single exit point
- All business logic is testable
- `defer` statements always execute

---

## Command-Line Interfaces

### Flag Naming

Follow the repository's convention. Without one, the pack default is
`snake_case`, matching Google's flag conventions and `go-packages/SKILL.md`:

```go
// Good: the pack default, used consistently
flag.String("output_dir", ".", "directory for output files")
flag.Bool("dry_run", false, "print actions without executing")

// Bad: two conventions in one binary
flag.String("outputDir", ".", "")    // camelCase
flag.String("output-dir", ".", "")   // hyphens beside snake_case flags
```

### Subcommands

For complex CLIs with subcommands, use `flag.NewFlagSet` per subcommand with
`flag.ContinueOnError`, so `run` returns the parse error and `main` stays the
only exit:

```go
func main() {
    if err := run(os.Args[1:]); err != nil && !errors.Is(err, flag.ErrHelp) {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(2)
    }
}

func run(args []string) error {
    if len(args) == 0 {
        return errors.New("usage: tool serve|migrate [flags]")
    }
    switch args[0] {
    case "serve":
        fs := flag.NewFlagSet("serve", flag.ContinueOnError)
        port := fs.Int("port", 8080, "listen port")
        if err := fs.Parse(args[1:]); err != nil {
            return err
        }
        return serve(*port)
    case "migrate":
        fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
        dryRun := fs.Bool("dry_run", false, "print changes without applying them")
        if err := fs.Parse(args[1:]); err != nil {
            return err
        }
        return migrate(*dryRun)
    default:
        return fmt.Errorf("unknown command %q; usage: tool serve|migrate [flags]", args[0])
    }
}
```

For larger CLIs, consider libraries like `cobra` or `urfave/cli`. Exit only from
`main()`.
