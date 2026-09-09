# All-skills review fixes

User approval: “исправь находки ревью”, referring to `docs/ALL_SKILLS_OPUS5_GPT56_REVIEW.md`.

Implement focused corrections in the shared Go 1.27 pack. Preserve the existing
Concision Gate, historical evidence, unrelated working-tree changes, and host
configuration. No commit, push, installation, or paid model benchmark is part
of this change. The audit supplies the design; no additional approval is needed.

- [x] Correct `go-security` JSON/SQL/parser and subprocess guidance; reuse the HTTP single-document owner.
- [x] Correct clone ownership, bounded concurrency examples/routing, context values, and zero/default/panic rules in skills and their references.
- [x] Align review-only actions, finding evidence, resource paths, and verification scope with `go-style-core`/`go-linting`.
- [x] Fix concrete reference drift and unjustified universal recommendations identified in the per-skill matrix; preserve useful domain contracts.
- [x] In `evals/cmd/abrun`, add failing regressions for completed refactor no-op and false skill-read attribution; fix the minimum responsible code and preserve implementation failure semantics.
- [x] Add executable regressions for bounded fan-out and contiguous-row append isolation; extend malformed JSON cases. Add distinct quality/trigger cases for the uncovered skill decisions and update published counts.
- [x] Run independent skill-application probes and review the final diff; distinguish these from model-specific A/B evidence.
- [x] Run all 24 skill validators, repository tests, race tests for changed Go code, build/vet, configured lint and link checks; refresh graft after code changes.
- [x] Record exact completed scope, checks, and any remaining evaluation limits in the audit report.

Checks run from `evals/`: `go test -count=1 ./...`,
`go test -race ./cmd/abrun`, `go build ./...`, `go vet ./...`.
Use `/Users/eugeneshershen/go/bin/golangci-lint`, built for Go 1.27.
Compile extracted Markdown examples through the existing `exampleBlock` and
`runExampleTest` helpers. Keep fixture-specific failures separate from schema,
tooling, and model behavior claims.
