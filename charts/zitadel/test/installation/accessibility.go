package installation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
)

type checkOptions struct {
	getUrl string
	test   func(response *http.Response) error
}

func (c *checkOptions) execute(ctx context.Context, t *testing.T, wg *sync.WaitGroup) {
	checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
	defer checkCancel()
	defer wg.Done()
	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, c.getUrl, nil)
	if err != nil {
		t.Fatalf("creating request for url %s failed: %s", c.getUrl, err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("sending request %+v failed: %s", *req, err)
		return
	}
	defer resp.Body.Close()

	if err = c.test(resp); err != nil {
		t.Fatalf("checking response to request %+v failed: %s", *req, err)
		return
	}
}

func (s *configurationTest) checkAccessibility(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.context, time.Minute)
	defer cancel()

	serviceTunnel := k8s.NewTunnel(s.kubeOptions, k8s.ResourceTypeService, s.zitadelRelease, 8080, 8080)
	serviceTunnel.ForwardPort(s.T())

	tunnels := []*k8s.Tunnel{serviceTunnel}
	defer func() {
		for _, t := range tunnels {
			t.Close()
		}
	}()

	checks := append(zitadelStatusChecks(8080), &checkOptions{
		getUrl: "http://localhost:8080/ui/console/assets/environment.json",
		test: func(resp *http.Response) error {
			if err := checkHttpStatus200(resp); err != nil {
				return err
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			err = nil
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

		podTunnel := k8s.NewTunnel(s.kubeOptions, k8s.ResourceTypePod, pod.Name, port, 8080)
		podTunnel.ForwardPort(s.T())
		tunnels = append(tunnels, podTunnel)
		checks = append(checks, zitadelStatusChecks(port)...)
	}

	wg := sync.WaitGroup{}
	for _, check := range checks {
		wg.Add(1)
		go check.execute(ctx, s.T(), &wg)
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

func checkHttpStatus200(resp *http.Response) error {
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected status code 200 but got %d", resp.StatusCode)
	}
	return nil
}
