# Third-Party Notices

This repository is Apache-2.0 for project-authored skills, scripts, docs,
assets, and evals. The `source/` directory contains bundled upstream source
snapshots that retain their upstream licenses. Each source file also includes an
inline provenance header with source URL, license, and copyright.

## Bundled Source Snapshots

| Path | Upstream project | Snapshot source | License | Copyright |
|------|------------------|-----------------|---------|-----------|
| `source/google-go-styleguide/guide.md` | Google Go Style Guide | `https://github.com/google/styleguide/blob/cc72fb0f41906a53d8dbf7cd0097aa8aea5e584b/go/guide.md` | CC-BY 3.0 | Google LLC |
| `source/google-go-styleguide/decisions.md` | Google Go Style Guide | `https://github.com/google/styleguide/blob/cc72fb0f41906a53d8dbf7cd0097aa8aea5e584b/go/decisions.md` | CC-BY 3.0 | Google LLC |
| `source/google-go-styleguide/best-practices.md` | Google Go Style Guide | `https://github.com/google/styleguide/blob/cc72fb0f41906a53d8dbf7cd0097aa8aea5e584b/go/best-practices.md` | CC-BY 3.0 | Google LLC |
| `source/effective-go/effective_go.html` | Go website, Effective Go | `https://github.com/golang/website/blob/b81d4dff74797e7ace4b63097cf5b38fa8388019/_content/doc/effective_go.html` | BSD 3-Clause | The Go Authors |
| `source/golang-wiki/CodeReviewComments.md` | Go Wiki, CodeReviewComments | `https://github.com/golang/wiki/blob/09ea82f9c67d4ced287d5e60dea11ddeda821bd3/CodeReviewComments.md` | BSD 3-Clause | The Go Authors |
| `source/uber-go-style/style.md` | Uber Go Style Guide | `https://github.com/uber-go/guide/blob/e2c8a0ed5723473c68e4deb28361bdc605ba8e98/style.md` | Apache-2.0 | Uber Technologies, Inc. |

Snapshot dates are not recorded separately in this repository. The immutable
commit IDs in the snapshot URLs are the source-of-truth provenance markers.

## Attribution

The Google Go Style Guide files are licensed under Creative Commons Attribution
3.0 and are attributed to Google LLC.

The Effective Go and Go Wiki CodeReviewComments snapshots are from the Go
project and are attributed to The Go Authors under the BSD 3-Clause license.

The Uber Go Style Guide snapshot is attributed to Uber Technologies, Inc. under
the Apache License, Version 2.0.

## Derived Guidance

No files from the projects below are bundled in this repository. They are
recorded because guidance in this repository was written after reviewing them
and follows their selection of topics closely enough to warrant attribution.

| This repository | Upstream project | License | Copyright | What was drawn from it |
|---|---|---|---|---|
| `skills/go-code-refactor/references/CATALOG.md`, `SAFETY-NET.md`, `MECHANICAL.md`, `STRUCTURAL.md`; the "When Not to Refactor" and "Risk Tiers" sections of `skills/go-code-refactor/SKILL.md` | [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang) `golang-refactoring` and `golang-design-patterns` skills | MIT | samber | Topic selection: the risk-tier split, coverage-adaptive safety net, bulk-rewrite tool ladder, type-alias gradual repair, import-cycle strategy order, and the not-to-refactor gate. Text is this repository's own. |
| `evals/ab/README.md` (the trap rule for a fixture) | [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang) skill eval format | MIT | samber | The practice of naming, per eval case, how a model fails *without* the skill |

## Provenance Policy

When updating files under `source/`:

1. Keep an inline header in every source snapshot with `Source`, `License`, and
   `Copyright`.
2. Use immutable upstream URLs when possible, preferably URLs pinned to a commit.
3. Update this notices file in the same change as any source snapshot addition,
   removal, or provenance change.
4. Treat `source/` as reference material. Synthesized skill guidance should cite
   the relevant source path or official Go documentation when the source affects
   a rule.
5. Verify version-sensitive standard-library guidance against official Go docs
   or release notes before changing skill recommendations.
