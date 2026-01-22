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

// CheckLogin performs a complete browser-based login flow verification against
// the ZITADEL instance. It uses headless Chrome via chromedp to simulate a real
// user logging in with the default admin credentials (Password1!), completing
// the mandatory password change on first login, optionally skipping MFA setup,
// and verifying that the user lands on the authenticated console page.
//
// The function creates its own browser context with a 5-minute timeout and
// captures screenshots on failure for debugging. Screenshots are saved to the
// .login-failures directory relative to the test working directory.
//
// This check validates the entire authentication pipeline including:
//   - OAuth/OIDC redirect flows
//   - Login UI rendering and interactivity
//   - Password validation and change workflows
//   - Session management and console access
func CheckLogin(t *testing.T, apiBaseURL string) {
	t.Helper()

	apiURL, err := url.Parse(apiBaseURL)
	require.NoError(t, err)

	loginFailuresDir := filepath.Join(".login-failures", t.Name())
	userDataDir, err := os.MkdirTemp("", "chromedp-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.RemoveAll(userDataDir); err != nil {
			t.Logf("Warning: failed to cleanup temp directory %s: %v", userDataDir, err)
		}
	})

	allocatorOpts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.IgnoreCertErrors,
		chromedp.NoSandbox,
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserDataDir(userDataDir),
		chromedp.WSURLReadTimeout(60*time.Second),
	)
	if execPath := findChromeExecutable(); execPath != "" {
		allocatorOpts = append(allocatorOpts, chromedp.ExecPath(execPath))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocatorOpts...)
	t.Cleanup(allocCancel)

	browserCtx, browserCancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(t.Logf),
		chromedp.WithErrorf(t.Logf),
	)
	t.Cleanup(browserCancel)

	loginCtx, loginCancel := context.WithTimeout(browserCtx, 5*time.Minute)
	defer loginCancel()

	err = chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		return nil
	}))
	require.NoError(t, err)

	loginHint := fmt.Sprintf("zitadel-admin@zitadel.%s", apiURL.Hostname())

	t.Run("navigate", func(t *testing.T) {
		runWithScreenshotOnFailure(t, loginCtx, loginFailuresDir, 60*time.Second,
			chromedp.Navigate(apiBaseURL+"/ui/console?login_hint="+loginHint),
		)
	})

	t.Run("await password page", func(t *testing.T) {
		runWithScreenshotOnFailure(t, loginCtx, loginFailuresDir, 30*time.Second,
			chromedp.WaitVisible(testIDSelector("password-text-input"), chromedp.ByQuery),
		)
	})

	t.Run("enter password", func(t *testing.T) {
		runWithScreenshotOnFailure(t, loginCtx, loginFailuresDir, 30*time.Second,
			chromedp.SendKeys(testIDSelector("password-text-input"), "Password1!", chromedp.ByQuery),
			chromedp.Click(testIDSelector("submit-button"), chromedp.ByQuery),
		)
	})

	t.Run("change password", func(t *testing.T) {
		var finalURL string
		runWithScreenshotOnFailure(t, loginCtx, loginFailuresDir, 60*time.Second,
			waitForPath("/ui/v2/login/password/change", 15*time.Second),
			chromedp.WaitVisible(testIDSelector("password-change-text-input"), chromedp.ByQuery),
			chromedp.WaitVisible(testIDSelector("password-change-confirm-text-input"), chromedp.ByQuery),
			chromedp.SendKeys(testIDSelector("password-change-text-input"), "Password2!", chromedp.ByQuery),
			chromedp.SendKeys(testIDSelector("password-change-confirm-text-input"), "Password2!", chromedp.ByQuery),
			chromedp.WaitEnabled(testIDSelector("submit-button"), chromedp.ByQuery),
			chromedp.Click(testIDSelector("submit-button"), chromedp.ByQuery),
			chromedp.Sleep(10*time.Second),
			chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Location(&finalURL).Do(ctx)
			}),
		)

		if strings.Contains(finalURL, "/ui/console") {
			t.Log("Redirected directly to console, skipping MFA test")
		} else {
			t.Run("skip mfa", func(t *testing.T) {
				runWithScreenshotOnFailure(t, loginCtx, loginFailuresDir, 30*time.Second,
					waitForPath("/ui/v2/login/mfa/set", 15*time.Second),
					chromedp.WaitVisible(testIDSelector("reset-button"), chromedp.ByQuery),
					chromedp.Click(testIDSelector("reset-button"), chromedp.ByQuery),
				)
			})
		}
	})

	t.Run("show console", func(t *testing.T) {
		runWithScreenshotOnFailure(t, loginCtx, loginFailuresDir, 60*time.Second,
			waitForPath("/ui/console", 15*time.Second),
			chromedp.WaitVisible("[data-e2e='authenticated-welcome']", chromedp.ByQuery),
		)
	})
}

func runWithScreenshotOnFailure(t *testing.T, ctx context.Context, failuresDir string, timeout time.Duration, actions ...chromedp.Action) {
	t.Helper()
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := chromedp.RunResponse(timeoutCtx, actions...)
	if err != nil {
		var (
			html            string
			screenshot      []byte
			debugFilePrefix = filepath.Join(failuresDir, strings.ReplaceAll(t.Name(), "/", "_")+"_fail")
		)
		_ = os.MkdirAll(failuresDir, os.ModePerm)
		_ = chromedp.Run(ctx, chromedp.OuterHTML("html", &html))
		_ = os.WriteFile(debugFilePrefix+".html", []byte(html), os.ModePerm)
		_ = chromedp.Run(ctx, chromedp.FullScreenshot(&screenshot, 100))
		_ = os.WriteFile(debugFilePrefix+".png", screenshot, os.ModePerm)
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

func testIDSelector(id string) string {
	return fmt.Sprintf("[data-testid='%s']", id)
}

func findChromeExecutable() string {
	candidates := []string{
		os.Getenv("CHROME_PATH"),
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/chrome",
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
