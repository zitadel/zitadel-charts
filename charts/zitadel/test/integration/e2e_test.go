package integration

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func (s *integrationTest) TestZITADELEnd2End() {
	// given
	options := &helm.Options{
		KubectlOptions: s.kubeOptions,
		SetStrValues: map[string]string{
			"zitadel.masterkey":                           "x123456789012345678901234567891y",
			"zitadel.secretConfig.Database.User.Password": "xy",
			"zitadel.configmapConfig.ExternalDomain":      "test.domain",
		},
	}

	// when
	helm.Install(s.T(), options, s.chartPath, s.release)

	// then
	// await that all zitadel related pods become ready
	pods := k8s.ListPods(s.T(), s.kubeOptions, metav1.ListOptions{LabelSelector: `app.kubernetes.io/instance=zitadel-test, app.kubernetes.io/component notin (init)`})
	s.awaitAvailability(pods)
	zitadelPods := make([]corev1.Pod, 0)
	for i := range pods {
		pod := pods[i]
		if name, ok := pod.GetObjectMeta().GetLabels()["app.kubernetes.io/name"]; ok && name == "zitadel" {
			zitadelPods = append(zitadelPods, pod)
		}
	}
	s.awaitListening(zitadelPods)
	s.awaitAccessibility(zitadelPods)
}

type checkOptions struct {
	getUrl string
	test   func(response *http.Response) error
}

func (s *integrationTest) awaitListening(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.context, 1*time.Minute)
	defer cancel()

	k8sClient, err := k8s.GetKubernetesClientE(s.T())
	if err != nil {
		s.T().Fatal(err)
	}

	podsClient := k8sClient.CoreV1().Pods(s.namespace)

	wg := sync.WaitGroup{}
	for _, pod := range pods {
		wg.Add(1)
		go awaitisteningMessage(s.context, s.T(), &wg, podsClient, pod)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				s.T().Fatalf("awaiting availability failed: %s", err)
			}
			return
		}
	}
}

func awaitisteningMessage(ctx context.Context, t *testing.T, wg *sync.WaitGroup, client v1.PodInterface, pod corev1.Pod) {
	reader, err := client.GetLogs(pod.Name, &corev1.PodLogOptions{Follow: true}).Stream(ctx)
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logLine := scanner.Text()
		if strings.Contains(logLine, "server is listening") {
			wg.Done()
			return
		}
		if err = scanner.Err(); err != nil {
			t.Fatal(err)
		}
	}
}

func (s *integrationTest) awaitAvailability(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.context, 5*time.Minute)
	defer cancel()

	wg := sync.WaitGroup{}
	for _, p := range pods {
		wg.Add(1)
		go func(pod corev1.Pod) {
			k8s.WaitUntilPodAvailable(s.T(), s.kubeOptions, pod.Name, 180, time.Second)
			wg.Done()
		}(p)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				s.T().Fatalf("awaiting availability failed: %s", err)
			}
			return
		}
	}
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

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				s.T().Fatalf("awaiting accessibility failed: %s", err)
			}
			return
		}
	}
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
