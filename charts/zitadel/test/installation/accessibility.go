package installation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
)

type checkOptions struct {
	getUrl string
	test   func(response *http.Response, body []byte) error
}

func (c *checkOptions) execute(ctx context.Context) (err error) {
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
	tunnels := []interface{ Close() }{CloseFunc(ServiceTunnel(s))}
	defer func() {
		for _, t := range tunnels {
			t.Close()
		}
	}()
	checks := append(zitadelStatusChecks(8080), &checkOptions{
		getUrl: "http://localhost:8080/ui/console/assets/environment.json",
		test: func(resp *http.Response, body []byte) error {
			if err := checkHttpStatus200(resp, body); err != nil {
				return err
			}
			bodyStr := string(body)
			for _, expect := range []string{
				`"api":"http://localhost:8080"`,
				`"issuer":"http://localhost:8080"`,
			} {
				if !strings.Contains(bodyStr, expect) {
					return fmt.Errorf("couldn't find %s in environment.json content %s", expect, bodyStr)
				}
			}
			return nil
		},
	})
	for i := range pods {
		pod := pods[i]
		port := 8081 + i

		podTunnel := k8s.NewTunnel(s.KubeOptions, k8s.ResourceTypePod, pod.Name, port, 8080)
		podTunnel.ForwardPort(s.T())
		tunnels = append(tunnels, podTunnel)
		checks = append(checks, zitadelStatusChecks(port)...)
	}
	wg := sync.WaitGroup{}
	for _, check := range checks {
		wg.Add(1)
		go Await(ctx, s.T(), &wg, 60, check.execute)
	}
	wait(ctx, s.T(), &wg, "accessibility")
}

func zitadelStatusChecks(port int) []*checkOptions {
	return []*checkOptions{{
		getUrl: fmt.Sprintf("http://localhost:%d/debug/validate", port),
		test:   checkHttpStatus200,
	}, {
		getUrl: fmt.Sprintf("http://localhost:%d/debug/healthz", port),
		test:   checkHttpStatus200,
	}, {
		getUrl: fmt.Sprintf("http://localhost:%d/debug/ready", port),
		test:   checkHttpStatus200,
	}}
}

func checkHttpStatus200(resp *http.Response, _ []byte) error {
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected status code 200 but got %d", resp.StatusCode)
	}
	return nil
}
