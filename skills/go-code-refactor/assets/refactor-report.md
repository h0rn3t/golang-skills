# Refactor Report Template

Copy this structure for the final answer. Write it in the language the user is
writing in — the headings below are English placeholders, not fixed strings.
Keep the prose short: the table and the diff carry the information, and a
paragraph restating them is noise.

---

## Summary

One paragraph: what scope was covered, and what the code now reads like that it
did not before. No feature tour. End with the instrument: `net: -<N> lines`
(and `-<M> deps` if any). A positive number needs a sentence of justification.

## Deleted

Dead code, unused parameters, redundant checks — lines removed and why each was
**provably** dead. Lead with this section: it is the part of the diff that
needed no design decision, and reviewers approve it at a glance.

## Changes

| Location | Before | After | Why |
|---|---|---|---|
| `file.go:120-260` | 140-line handler mixing decode, validation, persistence | orchestrator + 4 named steps | one job per function |

One row per named transformation. Terse — the diff shows the detail.

## Modernization

The `go` directive and toolchain version. What `go fix` did versus what you did
by hand. Which new APIs were adopted, and one clause each on why the swap is
behavior-identical. Anything from Tier 2 that needed a condition checked, and
what you checked.

## Behavior risk

Places where the refactor came close to changing semantics, and what keeps it
safe. Cross-reference the "Kept:" comments left in the code, so the reader can
find them without grepping.

## Verification

```
                 baseline    after
gofmt            pass        pass
build            pass        pass
vet              pass        pass
test             pass        pass   (42 tests, same names)
race             pass        pass
lint findings    17          4
diff             — empty —
```

Paste the real `verify-refactor.sh diff` result. **Report a skipped check as
skipped**, never as passed: if `golangci-lint` was not installed or a
build-tagged file could not be compiled here, say so on its own line.

## Findings — not applied

Bugs, races, ignored errors, and Tier 3 modernizations. One line of rationale
each, with a location. These were deliberately left out of the diff because
fixing them would have changed behavior; the user decides what happens next.

| Location | Finding | Why it was not fixed here |
|---|---|---|
| `store.go:88` | `err` discarded with no justifying comment | adding a check changes control flow |
| `api.go:31` | `omitempty` → `omitzero` would read better | changes the JSON on the wire |
