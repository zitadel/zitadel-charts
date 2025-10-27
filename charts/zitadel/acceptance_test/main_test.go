package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
)

type panicT struct {
	testing.T
}

func (p *panicT) Errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func TestMain(m *testing.M) {
	os.Exit(func() int {
		t := &panicT{}
		helm.AddRepo(t, &helm.Options{}, "traefik", "https://traefik.github.io/charts")
		_, filename, _, _ := runtime.Caller(0)
		traefikOptions := &helm.Options{
			Version:     "37.2.0",
			ValuesFiles: []string{filepath.Join(filename, "..", "..", "..", "..", "examples", "99-kind-with-traefik", "traefik-values.yaml")},
			ExtraArgs:   map[string][]string{"upgrade": {"--install", "--wait", "--namespace", "ingress", "--create-namespace"}},
		}
		helm.Upgrade(t, traefikOptions, "traefik/traefik", "traefik")
		return m.Run()
	}())
}
