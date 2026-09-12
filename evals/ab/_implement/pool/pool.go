// Package pool runs independent jobs on a bounded number of workers.
package pool

import "context"

// Job is one unit of work. A Job returns when its work is done or when ctx is
// cancelled, whichever comes first.
type Job func(ctx context.Context) error

// Run runs every job in jobs with at most workers of them in flight at once,
// and returns once the work is over.
//
// The jobs are the nightly imports of a data team: each one is independent
// of the others, each one holds a connection to a partner system while it
// runs, and the partner systems allow this process a fixed number of
// connections in total. That number is workers. Opening more connections
// than that is what got the previous importer's credentials revoked.
//
// The first job to fail ends the run: jobs that have not started are not
// started, the ctx passed to the jobs still running is cancelled, and Run
// returns that first error once every job it started has returned. Cancelling
// ctx ends the run the same way, with ctx.Err() as the result when no job
// failed first. When every job returns nil, Run returns nil.
//
// Run owns the goroutines it starts. When it returns, none of them is still
// running; the operator reads a goroutine dump after an import and expects to
// find this package absent from it.
//
// workers below 1 is an error whose message names workers. No jobs is a
// successful run.
func Run(ctx context.Context, workers int, jobs []Job) error {
	panic("not implemented")
}
