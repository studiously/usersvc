package introspector

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"strings"

	"github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc/metadata"
)

const (
	bearer       string = "bearer"
	bearerFormat string = "Bearer %s"
)

// ToHTTPContext moves OAuth2 token from request header to context. Particularly
// useful for servers.
func ToHTTPContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := extractTokenFromAuthHeader(r.Header.Get("Authorization"))
		if !ok {
			return ctx
		}
		return context.WithValue(ctx, TokenContextKey, token)
	}
}

// FromHTTPContext moves OAuth2 token from context to request header. Particularly useful for clients.
func FromHTTPContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := ctx.Value(TokenContextKey).(string)
		if ok {
			r.Header.Add("Authorization", generateAuthHeaderFromToken(token))
		}
		return ctx
	}
}

// ToGRPCContext moves OAuth2 token from gRPC metadata to context. Particularly useful for servers.
func ToGRPCContext() grpc.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		// capital "Key" is illegal in HTTP/2.
		authHeader, ok := md["authorization"]
		if !ok {
			return ctx
		}

		token, ok := extractTokenFromAuthHeader(authHeader[0])
		if ok {
			ctx = context.WithValue(ctx, TokenContextKey, token)
		}

		return ctx
	}
}

// FromGRPCContext moves OAuth2 token from context to gRPC metadata. Particularly useful for clients.
func FromGRPCContext() grpc.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		token, ok := ctx.Value(TokenContextKey).(string)
		if ok {
			// capital "Key" is illegal in HTTP/2.
			(*md)["authorization"] = []string{generateAuthHeaderFromToken(token)}
		}

		return ctx
	}
}

func extractTokenFromAuthHeader(val string) (token string, ok bool) {
	authHeaderParts := strings.Split(val, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != bearer {
		return "", false
	}

	return authHeaderParts[1], true
}

func generateAuthHeaderFromToken(token string) string {
	return fmt.Sprintf(bearerFormat, token)
}
