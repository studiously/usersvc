# introspector
Go kit middleware to introspect OAuth2 tokens with Hydra

Sample usage:
```go
r.Methods("GET").Path("/classes/{class}").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.get")(e.GetClassEndpoint),
		decodeGetClassRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(introspector.ToHTTPContext()))...
))
```