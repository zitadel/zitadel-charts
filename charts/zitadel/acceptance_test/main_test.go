package acceptance_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
)

// panicT is a test helper that converts test errors into panics, forcing
// immediate test failure during TestMain setup. This ensures that if the
// Traefik installation fails, the entire test suite stops immediately rather
// than running tests against a broken environment.
type panicT struct {
	testing.T
}

func (p *panicT) Errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

// TestMain sets up the test environment by installing Traefik ingress controller
// before running any acceptance tests. This ensures all tests have a functional
// ingress controller available for routing HTTP/HTTPS traffic to test deployments.
//
// The Traefik installation uses programmatic configuration matching the values from
// examples/99-kind-with-traefik/traefik-values.yaml, including NodePort service type
// with ports 30080 and 30443, debug logging, insecure TLS skip verification for
// self-signed certificates, and automatic HTTPS redirects.
//
// The function uses an IIFE (Immediately Invoked Function Expression) pattern to
// ensure proper cleanup and exit code handling regardless of setup success or failure.
func TestMain(m *testing.M) {
	os.Exit(func() int {
		t := &panicT{}

		helm.AddRepo(t, &helm.Options{}, "traefik", "https://traefik.github.io/charts")

		traefikOptions := &helm.Options{
			Version: "37.2.0",
			SetValues: map[string]string{
				"logs.general.level":                          "DEBUG",
				"additionalArguments[0]":                      "--serverstransport.insecureskipverify=true",
				"service.type":                                "NodePort",
				"ports.web.nodePort":                          "30080",
				"ports.websecure.nodePort":                    "30443",
				"ports.web.redirections.entryPoint.to":        "websecure",
				"ports.web.redirections.entryPoint.scheme":    "https",
				"ports.web.redirections.entryPoint.permanent": "true",
				"ingressClass.enabled":                        "true",
				"ingressClass.isDefaultClass":                 "true",
			},
			ExtraArgs: map[string][]string{
				"upgrade": {"--install", "--wait", "--namespace", "ingress", "--create-namespace"},
			},
		}

		helm.Upgrade(t, traefikOptions, "traefik/traefik", "traefik")
		return m.Run()
	}())
}
