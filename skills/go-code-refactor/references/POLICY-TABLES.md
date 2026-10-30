# Refactoring Shared Policy Tables

> Sources: project policy; source/google-go-styleguide/guide.md (Least mechanism)
> Authority: project policy
> Last verified: 2026-09-10

When several functions select fields of the same policy record, a shared
table can remove repeated selection:

```go
// before: the selection lives in every function, the literals twice over
func Rate(zone string) (int64, error) {
    if zone == "domestic" {
        return 499, nil
    } else if zone == "regional" {
        return 899, nil
    } else if zone == "overseas" {
        return 1899, nil
    }
    return 0, fmt.Errorf("rate %q: %w", zone, ErrUnknownZone)
}

func Surcharge(zone string, kg int) (int, error) {
    if zone == "domestic" {
        if kg >= 20 {
            return 10, nil
        }
        return 0, nil
    } else if zone == "regional" {
        if kg >= 20 {
            return 15, nil
        }
        return 0, nil
    } else if zone == "overseas" {
        if kg >= 20 {
            return 25, nil
        }
        return 0, nil
    }
    return 0, fmt.Errorf("surcharge %q: %w", zone, ErrUnknownZone)
}
```

```go
// after: one policy table; callers preserve their own errors
type zone struct {
    rate  int64
    heavy int // surcharge percent at 20 kg and above
}

var zones = map[string]zone{
    "domestic": {rate: 499, heavy: 10},
    "regional": {rate: 899, heavy: 15},
    "overseas": {rate: 1899, heavy: 25},
}

func Rate(name string) (int64, error) {
    z, ok := zones[name]
    if !ok {
        return 0, fmt.Errorf("rate %q: %w", name, ErrUnknownZone)
    }
    return z.rate, nil
}

func Surcharge(name string, kg int) (int, error) {
    z, ok := zones[name]
    if !ok {
        return 0, fmt.Errorf("surcharge %q: %w", name, ErrUnknownZone)
    }
    if kg >= 20 {
        return z.heavy, nil
    }
    return 0, nil
}
```

Here the zone facts live in one table, while the accessors retain their distinct
error behavior. The completion criteria and the choice of shape follow.

## Finishing the fold

Replacing literals with
constants or changing `if` to `switch` alone does not remove repeated logic.

Equal literals and matching switch keys do not establish a shared policy: a
retry limit of 3 and a grace period of 3 days stay independent facts. Combine
values only when they belong to one record and the final code reads better;
stop when another fold adds more coupling than it removes, and name a material
remaining duplication in the report.

Pick the shape by the final code, call sites included: an exported accessor
that already performs the selection, before a new unexported helper; a
`switch` when the cases carry logic; a `map` or slice literal indexed by the
key when they carry only values. A table exists to delete the branches, not
to be serviced — no search helper, no method, no loop to rebuild a list that
was already a literal. If the lookup needs those, the `switch` was shorter.
Map iteration order is not source order, so an ordered literal stays a
literal. Error texts and the point where an unknown key fails do not move.
