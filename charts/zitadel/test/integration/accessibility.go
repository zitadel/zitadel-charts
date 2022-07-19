package integration

import (
	"context"
	"errors"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type checkOptions struct {
	getUrl string
	test   func(response *http.Response) error
}

func (c *checkOptions) execute(t *testing.T, log *logger.Logger) bool {
	resp, err := http.Get(c.getUrl)
	if errors.Is(err, io.EOF) {
		log.Logf(t, "checking url %s: %s", c.getUrl, err)
		return false
	}
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if err := c.test(resp); err != nil {
		log.Logf(t, "checking url %s: %s", c.getUrl, err)
		return false
	}
	return true
}

func (s *integrationTest) awaitAccessibility(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.context, 5*time.Minute)
	defer cancel()

	k8s.NewTunnel(s.kubeOptions, k8s.ResourceTypeService, s.release, 8080, 8080).ForwardPort(s.T())

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
				`"api":"https://localhost:8080"`,
				`"issuer":"https://localhost:8080"`,
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

		k8s.NewTunnel(s.kubeOptions, k8s.ResourceTypePod, pod.Name, port, 8080).ForwardPort(s.T())
		checks = append(checks, zitadelStatusChecks(port)...)
	}

	wg := sync.WaitGroup{}
	for _, check := range checks {
		wg.Add(1)
		go executeChecks(s.T(), s.log, &wg, check)
	}
	wait(ctx, s.T(), &wg, "accessibility")
}

func executeChecks(t *testing.T, log *logger.Logger, wg *sync.WaitGroup, check *checkOptions) {
	for {
		if check.execute(t, log) {
			break
		}
		time.Sleep(time.Second)
	}
	wg.Done()
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
