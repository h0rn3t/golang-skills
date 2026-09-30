# Package Size, Program Structure, and CLIs

> Sources: source/google-go-styleguide/best-practices.md (Package size); source/uber-go-style/style.md (Avoid init(), Exit in Main); https://pkg.go.dev/flag
> Authority: advisory
> Last verified: 2026-09-10

## Contents

- [When to Split a Package](#when-to-split-a-package)
- [Avoiding init()](#avoiding-init)
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

---

## Avoiding init()

Prefer explicit functions over `init()`.

**Acceptable uses of init():**
- Complex expressions that cannot be single assignments
- Pluggable hooks (e.g., `database/sql` dialects, encoding registries)
- Deterministic precomputation

---

## Command-Line Interfaces

### Subcommands

For complex CLIs with subcommands, use `flag.NewFlagSet` per subcommand with
`flag.ContinueOnError`, so `run` returns the parse error and `main` stays the
only exit. `Parse` has already printed that error and the usage when it
returns, so `main` prints only the other errors, and exits 2 only for a usage
error:

```go
// errUsage marks a usage error already printed to stderr with the usage text.
var errUsage = errors.New("usage error")

func main() {
    err := run(os.Args[1:])
    switch {
    case err == nil, errors.Is(err, flag.ErrHelp):
    case errors.Is(err, errUsage):
        os.Exit(2)
    default:
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run(args []string) error {
    if len(args) == 0 {
        fmt.Fprintln(os.Stderr, "usage: tool serve|migrate [flags]")
        return errUsage
    }
    switch args[0] {
    case "serve":
        fs := flag.NewFlagSet("serve", flag.ContinueOnError)
        port := fs.Int("port", 8080, "listen port")
        if err := fs.Parse(args[1:]); err != nil {
            return fmt.Errorf("%w: %w", errUsage, err)
        }
        return serve(*port)
    case "migrate":
        fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
        dryRun := fs.Bool("dry_run", false, "print changes without applying them")
        if err := fs.Parse(args[1:]); err != nil {
            return fmt.Errorf("%w: %w", errUsage, err)
        }
        return migrate(*dryRun)
    default:
        fmt.Fprintf(os.Stderr, "unknown command %q\nusage: tool serve|migrate [flags]\n", args[0])
        return errUsage
    }
}
```

For larger CLIs, consider libraries like `cobra` or `urfave/cli`. Exit only from
`main()`.
