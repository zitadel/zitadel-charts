package acceptance_test

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
)

var CTX context.Context
var WithBrowserCTX func(ctx context.Context, t *testing.T) (context.Context, func())

type panicT struct {
	testing.T
}

func (p *panicT) Errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func TestMain(m *testing.M) {
	os.Exit(func() int {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		CTX = ctx
		execAllocatorOptions := append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.IgnoreCertErrors,
		)
		if chromeBinaryPath := os.Getenv("CHROME_BINARY_PATH"); chromeBinaryPath != "" {
			execAllocatorOptions = append(execAllocatorOptions, chromedp.ExecPath(chromeBinaryPath))
		}
		allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, execAllocatorOptions...)
		defer allocCancel()
		browserCtx, browserCancel := chromedp.NewContext(allocCtx)
		defer browserCancel()
		// a no-op action to allocate the browser
		if err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(_ context.Context) error {
			return nil
		})); err != nil {
			panic(err)
		}
		WithBrowserCTX = func(testCtx context.Context, t *testing.T) (context.Context, func()) {
			deadline, ok := testCtx.Deadline()
			if !ok {
				panic("deadline not set")
			}
			browserDeadlineCtx, browserDeadlineCancel := context.WithDeadline(browserCtx, deadline)
			browserChildCtx, _ := chromedp.NewContext(browserDeadlineCtx,
				chromedp.WithExistingBrowserContext(
					chromedp.FromContext(browserCtx).BrowserContextID,
				),
			)
			return browserChildCtx, browserDeadlineCancel
		}

		t := &panicT{}
		helm.AddRepo(t, &helm.Options{}, "traefik", "https://traefik.github.io/charts")
		_, filename, _, _ := runtime.Caller(0)
		traefikOptions := &helm.Options{
			Version:     "36.3.0",
			ValuesFiles: []string{filepath.Join(filename, "..", "..", "..", "..", "examples", "99-kind-with-traefik", "traefik-values.yaml")},
			ExtraArgs:   map[string][]string{"upgrade": {"--install", "--wait", "--namespace", "ingress", "--create-namespace"}},
		}
		helm.Upgrade(t, traefikOptions, "traefik/traefik", "traefik")
		return m.Run()
	}())
}
