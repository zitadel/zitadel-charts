package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// Get performs an HTTP GET request that skips TLS certificate verification.
// Returns the status code, response body, and any error encountered.
func Get(ctx context.Context, url string, headers map[string]string) (int, []byte, error) {
	return doRequest(ctx, http.MethodGet, url, headers, nil)
}

// Post performs an HTTP POST request that skips TLS certificate verification.
// Returns the status code, response body, and any error encountered.
func Post(ctx context.Context, url string, headers map[string]string, body io.Reader) (int, []byte, error) {
	return doRequest(ctx, http.MethodPost, url, headers, body)
}

func doRequest(ctx context.Context, method, url string, headers map[string]string, body io.Reader) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, fmt.Errorf("creating request failed: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := insecureClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("reading response failed: %w", err)
	}

	return resp.StatusCode, respBody, nil
}
