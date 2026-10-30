# Runtime Settings

> Sources: https://pkg.go.dev/runtime#GOMAXPROCS; https://go.dev/doc/gc-guide; https://go.dev/doc/pgo
> Authority: advisory; runtime defaults normative
> Minimum Go: cgroup-aware `GOMAXPROCS` 1.25
> Last verified: 2026-10-30

`GOMAXPROCS` follows the cgroup CPU limit from a `go 1.25` directive, so
`go.uber.org/automaxprocs` is a dependency to remove, as its own change
([MODERNIZATION.md](../../go-code-refactor/references/MODERNIZATION.md)). The
removal changes the value under a small or fractional limit: the runtime
rounds the limit up and stays at 2 or more on a machine with two CPUs or
more, where `automaxprocs` rounded down to at least 1 — a 1.5-CPU container
goes from 1 to 2. The runtime reads the CPU limit, not the request, and a
`GOMAXPROCS` variable or call turns its automatic updates off
(`go doc runtime.GOMAXPROCS`).

The garbage collector's settings belong to the process, not to a package:
library code never calls `debug.SetGCPercent`, `debug.SetMemoryLimit`,
`runtime.GOMAXPROCS`, or `runtime.GC`. A `GOGC` value is a deployment
recommendation, measured against the unchanged snapshot with
`--verdict DIR --new-env GOGC=200`. `GOMEMLIMIT` goes where the process owns
its container's memory, 5–10% under the limit, and never to avoid an
out-of-memory kill in a process already close to that limit: it trades the
kill for a severe slowdown, where a higher limit or a lower `GOGC` does not
([gc-guide](https://go.dev/doc/gc-guide)). Tuning chosen before Go 1.26 was
measured against the previous collector; re-measure it.

Profile-guided optimization is a build input, not a code change: a merged
CPU profile from production saved as `default.pgo` in the main package's
directory is used by `go build` (`-pgo=auto`, the default since Go 1.21),
typically for 2–14%. A microbenchmark's profile is a poor source: it
exercises one small part of the program
([go.dev/doc/pgo](https://go.dev/doc/pgo)).
