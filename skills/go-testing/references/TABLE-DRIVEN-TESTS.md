# Table-Driven Tests, Subtests, and Parallel Tests

> Sources: source/google-go-styleguide/decisions.md (Table-driven tests); source/uber-go-style/style.md (Test Tables)
> Authority: advisory
> Minimum Go: per-iteration loop variables 1.22
> Last verified: 2026-09-10

---

## Subtests

Use `t.Run` for better organization, filtering, and parallel execution.

### Subtest Names

- Use clear, concise names: `t.Run("empty_input", ...)`, `t.Run("hu_to_en", ...)`
- Avoid wordy descriptions or slashes (slashes break test filtering)
- Subtests must be independent — no shared state or execution order dependencies

### Table Tests with Subtests

```go
func TestTranslate(t *testing.T) {
    tests := []struct {
        name, srcLang, dstLang, input, want string
    }{
        {"hu_en_basic", "hu", "en", "köszönöm", "thank you"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Translate(tt.srcLang, tt.dstLang, tt.input); got != tt.want {
                t.Errorf("Translate(%q, %q, %q) = %q, want %q",
                    tt.srcLang, tt.dstLang, tt.input, got, tt.want)
            }
        })
    }
}
```

---

## Parallel Tests

Under a `go` directive of 1.22 or later a `tt := tt` in lines you touch is
dead code; `go fix -forvar <packages in scope>` removes it.
