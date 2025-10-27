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
	"github.com/hofstadter-io/cinful"
	"github.com/stretchr/testify/require"
)

// login performs a complete authentication flow including initial login,
// password change, optional MFA skip, and console verification. Uses a single
// consolidated loadPage call for the entire flow with appropriate timeouts
// and error handling.
func (s *ConfigurationTest) login(ctx context.Context, t *testing.T) {
	t.Helper()

	apiUrl, err := url.Parse(s.ApiBaseUrl)
	require.NoError(t, err, "Failed to parse API Base URL")

	loginURL := *apiUrl
	loginURL.Path = path.Join(loginURL.Path, "/ui/console")
	query := loginURL.Query()
	query.Set("login_hint", "zitadel-admin@zitadel."+apiUrl.Hostname())
	loginURL.RawQuery = query.Encode()

	userDataDir, err := os.MkdirTemp("", "chromedp-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.RemoveAll(userDataDir); err != nil {
			t.Logf("Failed to cleanup temp directory %s: %v", userDataDir, err)
		}
	})

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cinful.Info() != nil),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.NoSandbox,
		chromedp.Flag("incognito", true),
		chromedp.UserDataDir(userDataDir),
	)...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx, func() []chromedp.ContextOption {
		opts := []chromedp.ContextOption{
			chromedp.WithLogf(t.Logf),
			chromedp.WithErrorf(t.Logf),
		}
		if os.Getenv("DEBUG") != "" {
			opts = append(opts, chromedp.WithDebugf(t.Logf))
		}
		return opts
	}()...)
	defer browserCancel()

	loginCtx, loginCancel := context.WithTimeoutCause(browserCtx, 5*time.Minute, fmt.Errorf("login test timed out after 5 minutes"))
	defer loginCancel()

	time.Sleep(30 * time.Second)

	t.Logf("Starting login: %s", loginURL.String())

	var finalURL string

	loadPage(t, loginCtx, filepath.Join(".login-failures", s.KubeOptions.Namespace), 5*time.Minute,
		chromedp.Navigate(loginURL.String()),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(fmt.Sprintf(`[data-testid='%s']`, "password-text-input"), chromedp.ByQuery),
		chromedp.SendKeys(fmt.Sprintf(`[data-testid='%s']`, "password-text-input"), "Password1!", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.Click(fmt.Sprintf(`[data-testid='%s']`, "submit-button"), chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),

		waitForPath("/ui/v2/login/password/change", 5*time.Second),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(fmt.Sprintf(`[data-testid='%s']`, "password-change-text-input"), chromedp.ByQuery),
		chromedp.WaitVisible(fmt.Sprintf(`[data-testid='%s']`, "password-change-confirm-text-input"), chromedp.ByQuery),
		chromedp.SendKeys(fmt.Sprintf(`[data-testid='%s']`, "password-change-text-input"), "Password2!", chromedp.ByQuery),
		chromedp.SendKeys(fmt.Sprintf(`[data-testid='%s']`, "password-change-confirm-text-input"), "Password2!", chromedp.ByQuery),
		chromedp.WaitEnabled(fmt.Sprintf(`[data-testid='%s']`, "submit-button"), chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.Click(fmt.Sprintf(`[data-testid='%s']`, "submit-button"), chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),

		chromedp.ActionFunc(func(ctx context.Context) error {
			var currentURL string
			if err := chromedp.Location(&currentURL).Do(ctx); err != nil {
				return err
			}
			t.Logf("Current URL after password change: %s", currentURL)

			if strings.Contains(currentURL, "/ui/console") {
				t.Log("Redirected to console, skipping MFA")
			} else {
				t.Log("Skipping MFA setup")
				mfaActions := chromedp.Tasks{
					waitForPath("/ui/v2/login/mfa/set", 15*time.Second),
					chromedp.Sleep(1 * time.Millisecond),
					chromedp.WaitVisible(fmt.Sprintf(`[data-testid='%s']`, "reset-button"), chromedp.ByQuery),
					chromedp.Click(fmt.Sprintf(`[data-testid='%s']`, "reset-button"), chromedp.ByQuery),
				}
				if err := mfaActions.Do(ctx); err != nil {
					return fmt.Errorf("failed during MFA skip: %w", err)
				}
			}
			return nil
		}),

		waitForPath("/ui/console", 15*time.Second),
		chromedp.Sleep(1*time.Millisecond),
		chromedp.WaitVisible("[data-e2e='authenticated-welcome']", chromedp.ByQuery),

		chromedp.ActionFunc(func(ctx context.Context) error {
			time.Sleep(500 * time.Millisecond)
			return chromedp.Location(&finalURL).Do(ctx)
		}),
	)

	t.Logf("Login flow complete. Final URL: %s", finalURL)

	require.Contains(t, finalURL, "/ui/console", "Expected to land on console page after login")

	t.Log("Successfully authenticated and reached console")
}

// loadPage executes a sequence of ChromeDP actions with the specified timeout
// and captures debug information (HTML dump and screenshot) on failure. The
// debug artifacts are saved to the loginFailuresDir for troubleshooting.
func loadPage(t *testing.T, ctx context.Context, loginFailuresDir string, timeout time.Duration, actions ...chromedp.Action) {
	t.Helper()
	timeoutCtx, timeoutCancel := context.WithTimeoutCause(ctx, timeout, fmt.Errorf("test %s timed out after %s", t.Name(), timeout))
	defer timeoutCancel()
	_, err := chromedp.RunResponse(timeoutCtx, actions...)
	if err != nil {
		var (
			html            string
			screenshot      []byte
			parts           = strings.Split(t.Name(), "/")
			testName        = parts[len(parts)-1]
			debugFilePrefix = filepath.Join(loginFailuresDir, testName+"_fail_dump")
		)
		printErr(t, os.MkdirAll(loginFailuresDir, os.ModePerm))
		printErr(t, chromedp.Run(ctx, chromedp.OuterHTML("html", &html)))
		printErr(t, os.WriteFile(debugFilePrefix+".html", []byte(html), os.ModePerm))
		printErr(t, chromedp.Run(ctx, chromedp.FullScreenshot(&screenshot, 100)))
		printErr(t, os.WriteFile(debugFilePrefix+".png", screenshot, os.ModePerm))
	}
	require.NoError(t, err)
}

// waitForPath polls the current URL until it matches the expected path or
// the timeout expires. Returns an error if the expected path is not reached
// within the timeout duration.
func waitForPath(expectedPath string, timeout time.Duration) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		deadline := time.Now().Add(timeout)
		var currentURL string
		for time.Now().Before(deadline) {
			if err := chromedp.Location(&currentURL).Do(ctx); err != nil {
				return err
			}
			u, err := url.Parse(currentURL)
			if err != nil {
				return fmt.Errorf("failed to parse URL %s: %w", currentURL, err)
			}
			if path.Clean(u.Path) == path.Clean(expectedPath) {
				return nil
			}
			time.Sleep(250 * time.Millisecond)
		}
		return fmt.Errorf("timeout waiting for path %s", expectedPath)
	})
}

// printErr logs errors that occur during debug artifact creation without
// failing the test. This helper ensures test failures don't cascade when
// attempting to capture diagnostic information.
func printErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Logf("Could not create debug output: %v", err)
	}
}
