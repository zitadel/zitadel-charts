package acceptance_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	mgmt_api "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
)

type checkOptions interface {
	name() string
	execute(ctx context.Context) error
}

type httpCheckOptions struct {
	getUrl string
	test   func(response *http.Response, body []byte) error
}

func (c *httpCheckOptions) execute(ctx context.Context) (err error) {
	checkCtx, checkCancel := context.WithTimeoutCause(ctx, 5*time.Second, fmt.Errorf("http GET check %s timed out after 5 seconds", c.getUrl))
	defer checkCancel()
	//nolint:bodyclose
	resp, body, err := HttpGet(checkCtx, c.getUrl, nil)
	if err != nil {
		return fmt.Errorf("HttpGet failed with response %+v and body %+v: %w", resp, body, err)
	}
	if err = c.test(resp, body); err != nil {
		return fmt.Errorf("checking response %+v with body %+v failed: %w", resp, body, err)
	}
	return nil
}

func (c *httpCheckOptions) name() string {
	return "http GET " + c.getUrl
}

type grpcCheckOptions struct {
	checkName string
	test      func(ctx context.Context) error
}

func (c *grpcCheckOptions) execute(ctx context.Context) error {
	return c.test(ctx)
}

func (c *grpcCheckOptions) name() string {
	return c.checkName
}

func (s *ConfigurationTest) checkAccessibility(ctx context.Context, t *testing.T) {
	ctx, cancel := context.WithTimeoutCause(ctx, time.Minute, fmt.Errorf("accessibility checks timed out after a minute"))
	defer cancel()
	var checks = append(
		zitadelStatusChecks(s.ApiBaseUrl),
		&httpCheckOptions{
			getUrl: s.ApiBaseUrl + "/ui/console/assets/environment.json",
			test: func(resp *http.Response, body []byte) error {
				if err := checkHttpStatus200(resp, body); err != nil {
					return err
				}
				bodyStr := string(body)
				for _, expect := range []string{
					fmt.Sprintf(`"api":"%s"`, s.ApiBaseUrl),
					fmt.Sprintf(`"issuer":"%s"`, s.ApiBaseUrl),
				} {
					if !strings.Contains(bodyStr, expect) {
						return fmt.Errorf("couldn't find %s in environment.json content %s", expect, bodyStr)
					}
				}
				return nil
			},
		},
		&grpcCheckOptions{
			checkName: "zitadel",
			test: func(ctx context.Context) (err error) {
				conn, err := OpenGRPCConnection(ctx, s, nil)
				if err != nil {
					return fmt.Errorf("couldn't create gRPC management client: %w", err)
				}
				_, err = conn.Healthz(ctx, &mgmt_api.HealthzRequest{})
				return err
			},
		})
	for _, check := range checks {
		t.Run(check.name(), func(t *testing.T) {
			Awaitf(ctx, t, 1*time.Minute, check.execute, "check %s failed for a minute", check.name())
		})
	}
}

func zitadelStatusChecks(apiBaseURL string) []checkOptions {
	return []checkOptions{
		&httpCheckOptions{
			getUrl: fmt.Sprintf("%s/debug/validate", apiBaseURL),
			test:   checkHttpStatus200,
		},
		&httpCheckOptions{
			getUrl: fmt.Sprintf("%s/debug/healthz", apiBaseURL),
			test:   checkHttpStatus200,
		},
		&httpCheckOptions{
			getUrl: fmt.Sprintf("%s/debug/ready", apiBaseURL),
			test:   checkHttpStatus200,
		}}
}

func checkHttpStatus200(resp *http.Response, _ []byte) error {
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected status code 200 but got %d", resp.StatusCode)
	}
	return nil
}
