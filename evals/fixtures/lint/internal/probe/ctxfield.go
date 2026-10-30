package probe

import "context"

// Job keeps a context in a field: go-context passes it as a parameter.
type Job struct {
	Ctx  context.Context
	Name string
}

// Run takes the context where go-context wants it.
func Run(ctx context.Context, name string) error {
	return context.Cause(ctx)
}
