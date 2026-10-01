// go ./... skips names that begin with "_", and so does the script.
package draft

// Runner would be a finding if the script scanned this directory.
type Runner interface{ Run() error }

type job struct{}

func (job) Run() error { return nil }
