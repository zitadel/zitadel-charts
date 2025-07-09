package acceptance_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func (s *ConfigurationTest) login(t *testing.T) {
	apiUrl, err := url.Parse(s.ApiBaseUrl)
	loginFailuresDir := filepath.Join(".login-failures", s.KubeOptions.Namespace)
	require.NoError(t, err)
	ctxx, _ := chromedp.NewExecAllocator(context.Background(), append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.IgnoreCertErrors,
	)...)
	ctx, cancel := chromedp.NewContext(
		ctxx,
		chromedp.WithLogf(t.Logf),
		chromedp.WithDebugf(t.Logf),
		chromedp.WithErrorf(t.Logf),
	)
	defer cancel()
	t.Run("open browser", func(t *testing.T) {
		// We don't time out the first chromedp.Run context as this would stop the browser.
		// see chromedp.Run docs
		require.NoError(t, chromedp.Run(ctx, chromedp.Navigate(s.ApiBaseUrl+"/ui/console?login_hint=zitadel-admin@zitadel."+apiUrl.Hostname())))
	})
	t.Run("await password page", func(t *testing.T) {
		loadPage(t, ctx, loginFailuresDir, 10*time.Second,
			chromedp.WaitVisible(testIdSelector("password-text-input"), chromedp.ByQuery),
		)
	})
	t.Run("enter password", func(t *testing.T) {
		loadPage(t, ctx, loginFailuresDir, 10*time.Second,
			chromedp.SendKeys(testIdSelector("password-text-input"), "Password1!", chromedp.ByQuery),
			chromedp.Click(testIdSelector("submit-button"), chromedp.ByQuery),
		)
	})
	t.Run("change password", func(t *testing.T) {
		loadPage(t, ctx, loginFailuresDir, 10*time.Second,
			waitForPath("/ui/v2/login/password/change", 5*time.Second),
			chromedp.WaitVisible(testIdSelector("password-change-text-input"), chromedp.ByQuery),
			chromedp.WaitVisible(testIdSelector("password-change-confirm-text-input"), chromedp.ByQuery),
			chromedp.SendKeys(testIdSelector("password-change-text-input"), "Password2!", chromedp.ByQuery),
			chromedp.SendKeys(testIdSelector("password-change-confirm-text-input"), "Password2!", chromedp.ByQuery),
			chromedp.WaitEnabled(testIdSelector("submit-button"), chromedp.ByQuery),
			chromedp.Click(testIdSelector("submit-button"), chromedp.ByQuery),
		)
	})
	t.Run("skip mfa", func(t *testing.T) {
		loadPage(t, ctx, loginFailuresDir, 10*time.Second,
			waitForPath("/ui/v2/login/mfa/set", 5*time.Second),
			chromedp.WaitVisible(testIdSelector("reset-button"), chromedp.ByQuery),
			chromedp.Click(testIdSelector("reset-button"), chromedp.ByQuery),
		)
	})
	t.Run("show console", func(t *testing.T) {
		loadPage(t, ctx, loginFailuresDir, 10*time.Second,
			waitForPath("/ui/console", 5*time.Second),
			chromedp.WaitVisible("[data-e2e='authenticated-welcome'", chromedp.ByQuery),
		)
	})
}

func loadPage(t *testing.T, ctx context.Context, loginFailuresDir string, timeout time.Duration, actions ...chromedp.Action) {
	t.Helper()
	loadCtx, loadCancel := context.WithTimeout(ctx, timeout)
	defer loadCancel()
	_, err := chromedp.RunResponse(loadCtx, actions...)
	if err != nil {
		var (
			html            string
			screenshot      []byte
			debugFilePrefix = filepath.Join(loginFailuresDir, strings.Split(t.Name(), "/")[1]+"_fail_dump")
		)
		printErr(t, os.MkdirAll(loginFailuresDir, os.ModePerm))
		printErr(t, chromedp.Run(ctx, chromedp.OuterHTML("html", &html)))
		printErr(t, os.WriteFile(debugFilePrefix+".html", []byte(html), os.ModePerm))
		printErr(t, chromedp.Run(ctx, chromedp.FullScreenshot(&screenshot, 100)))
		printErr(t, os.WriteFile(debugFilePrefix+".png", screenshot, os.ModePerm))
	}
	require.NoError(t, err)
}

func waitForPath(expectedPath string, timeout time.Duration) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		deadline := time.Now().Add(timeout)
		var currentURL string
		for time.Now().Before(deadline) {
			if err := chromedp.Location(&currentURL).Do(ctx); err != nil {
				return err
			}
			u, _ := url.Parse(currentURL)
			if path.Clean(u.Path) == path.Clean(expectedPath) {
				return nil
			}
			time.Sleep(250 * time.Millisecond)
		}
		return fmt.Errorf("timeout waiting for path %s", expectedPath)
	})
}

func printErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Logf("⚠️ Could not create debug output for failed test: %v", err)
	}
}

type testIdSelector string

func (s testIdSelector) String() string {
	return fmt.Sprintf(`[data-testid='%s']`, string(s))
}
