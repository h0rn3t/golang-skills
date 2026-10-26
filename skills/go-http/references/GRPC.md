# gRPC Servers and Clients

> Sources: `google.golang.org/grpc` v1.84 package docs and source (`server.go`, `clientconn.go`, `status`, `health`, `test/bufconn`); protoc-gen-go-grpc generated code
> Authority: advisory (module forms, not stdlib)
> Minimum Go: the module's own floor; nothing here is version-gated by Go
> Last verified: 2026-10-26 against grpc v1.84.0

A package already on gRPC keeps its generated code and interceptor chain; the
rules below are what the generated code does not enforce. Authorization on
interceptors, streams, and peer identity is
[go-security](../../go-security/references/ACCESS-CONTROL.md#service-identity)'s.

## Errors cross the wire as a status

A handler returns `status.Error(codes.X, msg)`. Any other error reaches the
client as `codes.Unknown` with its `Error()` text, except `context.Canceled`
and `context.DeadlineExceeded`, which map to their codes. Map domain errors
once, at the handler, with `errors.Is`: not found → `NotFound`, a duplicate →
`AlreadyExists`, bad input → `InvalidArgument`, a state conflict →
`FailedPrecondition`, a dependency down → `Unavailable`, which clients retry.
An unexpected failure is `codes.Internal` with generic text and the cause in
the server log: `status.Errorf(codes.Internal, "lookup: %v", err)` sends the
database error to the client. The client reads the code with
`status.Code(err)`, which also sees a status wrapped with `%w`.

## Generated servers embed by value

Embed `pb.UnimplementedFooServer` by value, never `*pb.UnimplementedFooServer`:
a method added to the proto then answers `Unimplemented` instead of a compile
error at every server, and the generated `RegisterFooServer` panics at
startup on a nil pointer embed.

## Clients

`grpc.NewClient(target, opts...)`; `grpc.Dial` and `DialContext` are
deprecated. `NewClient` does not connect until the first call, ignores
`WithBlock` and `WithTimeout`, and resolves a bare `host:port` through `dns`
where `Dial` used `passthrough`: a test target that is not a resolvable name
takes `passthrough:///` explicitly. One `*grpc.ClientConn` per target, shared
by every goroutine and closed at shutdown, never one per request. Every call
carries its own deadline (`context.WithTimeout`); the server sees it as its
`ctx` deadline, so a handler that calls onward passes `ctx` on.

## Shutdown

`GracefulStop` blocks until every RPC finishes, and an open stream never
does. The fragment below bounds it and falls back to `Stop`, which closes
every connection:

```go
health.Shutdown() // NOT_SERVING first, so balancers drain this instance
done := make(chan struct{})
go func() { srv.GracefulStop(); close(done) }()
select {
case <-done:
case <-time.After(10 * time.Second):
	srv.Stop()
}
```

`health` is the `google.golang.org/grpc/health` server registered beside the
service; `SetServingStatus` reports one service, `Shutdown` all of them.

## Testing

`test/bufconn` serves the real server in memory, interceptors included. The
fragment below is the test setup:

```go
lis := bufconn.Listen(1 << 20)
srv := grpc.NewServer()
pb.RegisterFooServer(srv, newServer(store))
go srv.Serve(lis)
t.Cleanup(srv.Stop)

conn, err := grpc.NewClient("passthrough:///bufnet",
	grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}),
	grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
	t.Fatal(err)
}
t.Cleanup(func() { conn.Close() })
client := pb.NewFooClient(conn)
```

Assert on `status.Code(err)`, never on the error text, and add a stream case
that stays open past the shutdown deadline when the code owns shutdown.
