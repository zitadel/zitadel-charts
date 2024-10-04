package acceptance

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

func HttpGet(ctx context.Context, url string, beforeSend func(req *http.Request)) (*http.Response, []byte, error) {
	return httpCall(ctx, http.MethodGet, url, beforeSend, nil)
}

func HttpPost(ctx context.Context, url string, beforeSend func(req *http.Request), body io.Reader) (*http.Response, []byte, error) {
	return httpCall(ctx, http.MethodPost, url, beforeSend, body)
}

func httpCall(ctx context.Context, method string, url string, beforeSend func(req *http.Request), requestBody io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request for url %s failed: %s", url, err.Error())
	}
	if beforeSend != nil {
		beforeSend(req)
	}
	httpClient := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("sending request %+v failed: %s", *req, err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	return resp, responseBody, err
}
