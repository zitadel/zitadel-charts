package acceptance_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	mgmt_api "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
)

type checkOptions interface {
	execute(ctx context.Context) error
}

type checkOptionsFunc func(ctx context.Context) error

func (f checkOptionsFunc) execute(ctx context.Context) error {
	return f(ctx)
}

type httpCheckOptions struct {
	getUrl string
	test   func(response *http.Response, body []byte) error
}

func (c *httpCheckOptions) execute(ctx context.Context) (err error) {
	checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
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

func (s *ConfigurationTest) checkAccessibility() {
	ctx, cancel := context.WithTimeout(CTX, time.Minute)
	defer cancel()
	checks := append(
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
		checkOptionsFunc(func(ctx context.Context) error {
			conn, err := OpenGRPCConnection(s, nil)
			if err != nil {
				return fmt.Errorf("couldn't create gRPC management client: %w", err)
			}
			_, err = conn.Healthz(ctx, &mgmt_api.HealthzRequest{})
			return err
		}))

	wg := sync.WaitGroup{}
	for _, check := range checks {
		wg.Add(1)
		go Await(ctx, s.T(), &wg, 60, check.execute)
	}
	wait(ctx, s.T(), &wg, "accessibility")
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
