package acceptance

import (
	"context"
	"fmt"
	mgmt_api "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/management"
	"net/http"
	"strings"
	"sync"
	"time"

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
			randomInvalidKey := `{"type":"serviceaccount","keyId":"229185755715993707","key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAm8bpVfzWJuZsEz1VfTrwSAdkbH+i/u2NS4dv60lwIjtXzrU7\n1xZkHw9jxqz+c+APTaTzp1KY49Dc/wcwXv032FuD1GK2ZSRnMaHm8QnNt8Xhi0e8\nBlu3QQmlqxWPCI67wDPUwXoSHM+r9gQXn2pOR0oonoLP+Gzef+RRj1zUFpZmHWPX\nxw4UWWHwl4xChw9iyO4HbZZGe6wBVYVWe2BnvviCVEeKapyjaCqokZES38S4+S2X\nit202xLlRDyXs3XFWmBzHGmEsxx3LZZor85Kbph/bGjDcV8rdQC1YKC++z8OhuLp\n79GltP7YWrfMN3Z8iRUJQY9APrKQYtljVkWrnQIDAQABAoIBAQCIRZrLyRHCF+LF\ndes6UPvv1t+n9oQtRLxNLV7f0m+Q0p7+yhZeE01kyn67R4yU65YXk0w+vIfZC1a4\nlp5fCl73Gx+ZBP2QPyczCPHRPIVE1Yt33zoByevmrjzKDGMC1nIyMmVVF6eOorFI\n1s2ffEycGqir+b1bEkoWUTJ0Gn3Cf1PE4vTgenHhCrYSvMsbmszQ5GDlfxNj27qf\nF2YrnLx11GplMYU0YEzGqSQHxw76rrmF7yiTvbB+olsjXWARAJxBriSlrF2BDYQk\n+HJ8MEwhWhncaZH1i0Xz/jarDBizpo2o1+K1ZqF6RBUknT72EPnMxI9JsvS4FH44\nZfbrujBhAoGBAMQnx6tO79GpnBIAr7iELyUu5F4mCdU6D0rOAiCjXPpCUAdCDuwX\nzROonIGXPPmhzXXtxebeTz4cf+P8p6tUnrqpl/f0Oi1DMOzv0jL/SAUDC9uUrg6k\nurXZT2dgeONwd1pADyNXSpbZfwRE5IoecFg6cgFi4kune0mdG3mr8QjpAoGBAMtN\nerrMc+4bc3GsmWG4FSXn3xlWMeVGIo2/owP2P5MuMu0ibjofZkl28y0xo8dJgWmv\nLiFSEOhUy+TXZK7K1a2+fD+AXHHaHkBjNbTmCaAbf7rZnuUL4iZVpQyIoTCVuAwo\nC6bsE4TcwGddk4yZj/WZ7v1be+uNgeYwQr2UshyVAoGAN8pYsBCzhR6IlVY8pG50\nOk8sBNss0MjCsLQHRuEwAL37pRTUybG7UmwSl4k8foPWvEP0lcWFJFVWyrGBvulC\nfDTgVFXSdi02LS3Iy1hwU3yaUsnm96NCt5YnT2/Q8l96kuDFbXfWbzFNPxmZJu+h\nZHa7FknZs0rfdgCJYAHXfIECgYEAw3kSqSrNyMICJOkkbO2W/+RLAUx8GwttS8dX\nkQaip/wCoTi6rQ3lxnslY23YIFRPpvL1srn6YbiudrCXMOz7uNtvEYt01082SQha\n6j1IQfZOwLRfb7EWV29/i2aPPWynEqEqWuuf9N5f7MLvjH9WCHpibJ4aryhXHqGG\nekvPWWUCgYA5qDsPk5ykRWEALbunzB/RkpxR6LTLSwriU/OzRswOiKo8UPqH4JZI\nOsFAgudG5H+UOEGMuaSvIq0PLbGex16PjKqUsRwgIoPdH8183f9fxZSJDmr7ELIy\nZJEvE3eJnYwMOpSEZS0VR5Sw0CmKV2Hhd+u6rRB8YjXMP0nAVg8eOA==\n-----END RSA PRIVATE KEY-----\n","userId":"229185755715600491"}`
			conn, err := OpenGRPCConnection(s, []byte(randomInvalidKey))
			if err != nil {
				return fmt.Errorf("couldn't create gRPC management client: %w", err)
			}
			_, err = conn.Healthz(ctx, &mgmt_api.HealthzRequest{})
			// TODO: Why is the key checked on the healthz RPC?
			if strings.Contains(err.Error(), "Errors.AuthNKey.NotFound") {
				err = nil
			}
			return err
		}))
	for i := range pods {
		pod := pods[i]
		port := 8081 + i

		podTunnel := k8s.NewTunnel(s.KubeOptions, k8s.ResourceTypePod, pod.Name, port, 8080)
		podTunnel.ForwardPort(s.T())
		tunnels = append(tunnels, podTunnel)
		checks = append(checks, zitadelStatusChecks(s.Scheme, s.Domain, uint16(port))...)
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
