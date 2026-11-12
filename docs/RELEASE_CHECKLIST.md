# Release Checklist

Use this checklist before tagging a release.

## Documentation

- Update `CHANGELOG.md` and move relevant entries from `[Unreleased]` to the
  release version.
- Confirm `README.md` install, project structure, license, and provenance
  sections still match the repository.
- Confirm `THIRD_PARTY_NOTICES.md` covers every file under `source/`.

## Validation

Run the release validation commands from the repository root:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0

(cd evals && go test -count=1 -race -shuffle=on ./...)
(cd hooks/helper && go test -count=1 -race -shuffle=on ./...)

bash -n hooks/*.sh

golangci-lint config verify --config skills/go-linting/assets/golangci.yml
```

## Versioning

- Update plugin and marketplace version metadata when cutting a release.
  Keep `.claude-plugin/plugin.json`, `.codex-plugin/plugin.json`,
  `.cursor-plugin/plugin.json`, and the Claude marketplace versions aligned.
  The Codex and Cursor catalogs resolve the version from their plugin manifest.
- Use a `vX.Y.Z` tag.
- Push the tag and confirm the `Validate Skills` workflow passes for the tag.
  Its `release` job then creates the GitHub release from the tag's
  `CHANGELOG.md` section; a tag with no section fails that job.
