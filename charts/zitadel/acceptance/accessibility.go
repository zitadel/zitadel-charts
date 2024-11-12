package acceptance

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	mgmt_api "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"

	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
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

func (s *ConfigurationTest) checkAccessibility(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.Ctx, time.Minute)
	defer cancel()
	apiBaseURL := s.APIBaseURL()
	tunnels := []interface{ Close() }{CloseFunc(ServiceTunnel(s))}
	defer func() {
		for _, t := range tunnels {
			t.Close()
		}
	}()
	checks := append(
		zitadelStatusChecks(s.Scheme, s.Domain, s.Port),
		&httpCheckOptions{
			getUrl: apiBaseURL + "/ui/console/assets/environment.json",
			test: func(resp *http.Response, body []byte) error {
				if err := checkHttpStatus200(resp, body); err != nil {
					return err
				}
				bodyStr := string(body)
				for _, expect := range []string{
					fmt.Sprintf(`"api":"%s"`, apiBaseURL),
					fmt.Sprintf(`"issuer":"%s"`, apiBaseURL),
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
	for i := range pods {
		pod := pods[i]
		podTunnel := k8s.NewTunnel(s.KubeOptions, k8s.ResourceTypePod, pod.Name, 0, 8080)
		podTunnel.ForwardPort(s.T())
		tunnels = append(tunnels, podTunnel)
		localPort, err := strconv.ParseUint(strings.Split(podTunnel.Endpoint(), ":")[1], 10, 16)
		if err != nil {
			s.T().Fatal(err)
		}
		checks = append(checks, zitadelStatusChecks(s.Scheme, s.Domain, uint16(localPort))...)
	}
	wg := sync.WaitGroup{}
	for _, check := range checks {
		wg.Add(1)
		go Await(ctx, s.T(), &wg, 60, check.execute)
	}
	wait(ctx, s.T(), &wg, "accessibility")
}

func zitadelStatusChecks(scheme, domain string, port uint16) []checkOptions {
	return []checkOptions{
		&httpCheckOptions{
			getUrl: fmt.Sprintf("%s://%s:%d/debug/validate", scheme, domain, port),
			test:   checkHttpStatus200,
		},
		&httpCheckOptions{
			getUrl: fmt.Sprintf("%s://%s:%d/debug/healthz", scheme, domain, port),
			test:   checkHttpStatus200,
		},
		&httpCheckOptions{
			getUrl: fmt.Sprintf("%s://%s:%d/debug/ready", scheme, domain, port),
			test:   checkHttpStatus200,
		}}
}

func checkHttpStatus200(resp *http.Response, _ []byte) error {
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected status code 200 but got %d", resp.StatusCode)
	}
	return nil
}
