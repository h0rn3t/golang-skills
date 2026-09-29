package middleware

// Handler is consumed by Logged, so returning it is not the producer-owned
// case.
type Handler interface {
	Serve(req string) string
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(string) string

// Serve calls f.
func (f HandlerFunc) Serve(req string) string { return f(req) }

// Logged wraps h.
func Logged(h Handler) Handler {
	return HandlerFunc(func(req string) string { return h.Serve(req) })
}
