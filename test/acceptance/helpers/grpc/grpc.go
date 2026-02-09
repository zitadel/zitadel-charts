package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Dial creates a gRPC connection to the specified URL. The connection skips
// TLS certificate verification for testing against self-signed certificates.
func Dial(_ context.Context, apiBaseURL string) (*grpc.ClientConn, error) {
	parsedURL, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	host := parsedURL.Host
	if parsedURL.Port() == "" {
		if parsedURL.Scheme == "https" {
			host = parsedURL.Host + ":443"
		} else {
			host = parsedURL.Host + ":80"
		}
	}

	var creds credentials.TransportCredentials
	if parsedURL.Scheme == "https" {
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	} else {
		creds = insecure.NewCredentials()
	}

	return grpc.NewClient(host, grpc.WithTransportCredentials(creds))
}

// WithBearerToken adds a bearer token to the context for gRPC authentication.
func WithBearerToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

// Invoke calls a gRPC method with the given request and response messages.
func Invoke(ctx context.Context, conn *grpc.ClientConn, method string, req, reply interface{}) error {
	return conn.Invoke(ctx, method, req, reply)
}
