package acceptance_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type masterkeyExpected struct {
	secretCreated      bool
	secretName         string
	masterkeyGenerated bool
	masterkeyValue     string
	immutable          bool
	shouldFail         bool
	isExternal         bool
}

type testCase struct {
	name      string
	setValues map[string]string
	expected  masterkeyExpected
	setupFn   func(t *testing.T, opts *k8s.KubectlOptions) map[string]string
}

type MasterkeyTestSuite struct {
	suite.Suite
	chartPath string
}

func (s *MasterkeyTestSuite) SetupSuite() {
	chartPath, err := filepath.Abs("..")
	require.NoError(s.T(), err)
	s.chartPath = chartPath
}

//func (s *MasterkeyTestSuite) TestAutoGenerateMasterkey() {
//	s.T().Parallel()
//	s.runTestCase(testCase{
//		name:      "auto-generate-masterkey",
//		setValues: map[string]string{},
//		expected: masterkeyExpected{
//			secretCreated:      true,
//			secretName:         "",
//			masterkeyGenerated: true,
//			immutable:          true,
//		},
//	})
//}

func (s *MasterkeyTestSuite) TestExplicitMasterkeyValue() {
	s.T().Parallel()
	s.runTestCase(testCase{
		name: "explicit-masterkey-value",
		setValues: map[string]string{
			"zitadel.masterkey": "abcd1234efgh5678ijkl9012mnop3456",
		},
		expected: masterkeyExpected{
			secretCreated:      true,
			secretName:         "",
			masterkeyValue:     "abcd1234efgh5678ijkl9012mnop3456",
			masterkeyGenerated: false,
			immutable:          true,
		},
	})
}

func (s *MasterkeyTestSuite) TestExternalSecretReference() {
	s.T().Parallel()
	s.runTestCase(testCase{
		name:    "external-secret-reference",
		setupFn: setupExternalMasterkeySecret,
		expected: masterkeyExpected{
			secretCreated: true,
			secretName:    "my-external-masterkey",
			isExternal:    true,
		},
	})
}

func (s *MasterkeyTestSuite) TestBothMasterkeyAndSecretNameShouldFail() {
	s.T().Parallel()
	s.runTestCase(testCase{
		name: "both-set-should-fail",
		setValues: map[string]string{
			"zitadel.masterkey":           "abcd1234efgh5678ijkl9012mnop3456",
			"zitadel.masterkeySecretName": "my-external-masterkey",
		},
		expected: masterkeyExpected{
			secretCreated: false,
			shouldFail:    true,
		},
	})
}

func (s *MasterkeyTestSuite) runTestCase(tc testCase) {
	t := s.T()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	namespaceName := fmt.Sprintf("zitadel-masterkey-%s",
		sanitizeName(tc.name))
	kubeOptions := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(t, kubeOptions, namespaceName)

	deployPostgres(t, kubeOptions)

	externalDomain := fmt.Sprintf("masterkey-%s.127.0.0.1.sslip.io",
		sanitizeName(tc.name))
	apiBaseUrl := fmt.Sprintf("https://%s", externalDomain)

	commonSetValues := map[string]string{
		"ingress.enabled":                                          "true",
		"ingress.className":                                        "traefik",
		"ingress.annotations":                                      "null",
		"login.ingress.enabled":                                    "true",
		"login.ingress.className":                                  "traefik",
		"login.ingress.annotations":                                "null",
		"zitadel.configmapConfig.ExternalDomain":                   externalDomain,
		"zitadel.configmapConfig.ExternalPort":                     "443",
		"zitadel.configmapConfig.ExternalSecure":                   "true",
		"zitadel.configmapConfig.TLS.Enabled":                      "false",
		"zitadel.configmapConfig.Database.Postgres.Host":           "db-postgresql",
		"zitadel.configmapConfig.Database.Postgres.Port":           "5432",
		"zitadel.configmapConfig.Database.Postgres.Database":       "zitadel",
		"zitadel.configmapConfig.Database.Postgres.User.Username":  "postgres",
		"zitadel.configmapConfig.Database.Postgres.User.Password":  "postgres",
		"zitadel.configmapConfig.Database.Postgres.User.SSL.Mode":  "disable",
		"zitadel.configmapConfig.Database.Postgres.Admin.Username": "postgres",
		"zitadel.configmapConfig.Database.Postgres.Admin.Password": "postgres",
		"zitadel.configmapConfig.Database.Postgres.Admin.SSL.Mode": "disable",
		"zitadel.setupJob.activeDeadlineSeconds":                   "900",
	}

	mergedSetValues := make(map[string]string)
	for k, v := range commonSetValues {
		mergedSetValues[k] = v
	}
	for k, v := range tc.setValues {
		mergedSetValues[k] = v
	}

	if tc.setupFn != nil {
		additionalValues := tc.setupFn(t, kubeOptions)
		for k, v := range additionalValues {
			mergedSetValues[k] = v
		}
	}

	releaseName := fmt.Sprintf("zitadel-masterkey-%s",
		sanitizeName(tc.name))

	helmOptions := &helm.Options{
		KubectlOptions: kubeOptions,
		SetValues:      mergedSetValues,
		ExtraArgs: map[string][]string{
			"install": {"--timeout", "15m", "--hide-notes"},
		},
	}

	if tc.expected.shouldFail {
		err := helm.InstallE(t, helmOptions, s.chartPath, releaseName)
		require.Error(t, err)
		require.Contains(t, err.Error(),
			"set either .Values.zitadel.masterkey xor "+
				".Values.zitadel.masterkeySecretName")
		k8s.DeleteNamespace(t, kubeOptions, namespaceName)
		return
	}

	helm.Install(t, helmOptions, s.chartPath, releaseName)

	defer func() {
		if !t.Failed() {
			helm.Delete(t, helmOptions, releaseName, true)
			k8s.DeleteNamespace(t, kubeOptions, namespaceName)
		} else {
			t.Logf("Test failed - namespace %s and release %s "+
				"left for debugging", namespaceName, releaseName)
		}
	}()

	setupJobName := releaseName + "-setup"
	k8s.WaitUntilJobSucceed(t, kubeOptions, setupJobName, 900,
		1*time.Second)

	k8s.WaitUntilDeploymentAvailable(t, kubeOptions, releaseName,
		200, 5*time.Second)

	t.Logf("Waiting for ingress to be ready at %s", apiBaseUrl)
	waitForIngressReady(t, apiBaseUrl, 2*time.Minute)

	secretName := releaseName + "-masterkey"
	if tc.expected.secretName != "" {
		secretName = tc.expected.secretName
	}

	assertMasterkeySecret(t, kubeOptions, secretName, tc.expected)

	cfg := &ConfigurationTest{
		KubeOptions: kubeOptions,
		ApiBaseUrl:  apiBaseUrl,
	}
	cfg.SetT(t)
	cfg.login(ctx, t)
	assertGRPCWorks(ctx, t, cfg, "iam-admin")
}

func waitForIngressReady(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	t.Logf("=== Starting ingress readiness check ===")
	t.Logf("Target URL: %s", url)
	t.Logf("Timeout: %v", timeout)

	deadline := time.Now().Add(timeout)
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			t.Logf("Following redirect to: %s", req.URL.String())
			return nil // Follow redirects
		},
	}

	attemptNum := 0
	for time.Now().Before(deadline) {
		attemptNum++
		timeRemaining := time.Until(deadline)
		t.Logf("[Attempt %d] Checking %s/... (time remaining: %v)", attemptNum, url, timeRemaining)

		resp, err := client.Get(url + "/")
		if err == nil {
			statusCode := resp.StatusCode
			resp.Body.Close()

			t.Logf("[Attempt %d] Got response! Status code: %d", attemptNum, statusCode)

			// Only succeed on 2xx or 3xx status codes
			if statusCode >= 200 && statusCode < 400 {
				t.Logf("✓ Ingress is ready! Status: %d", statusCode)
				t.Logf("Adding 5 second buffer before proceeding...")
				time.Sleep(5 * time.Second)
				t.Logf("=== Ingress readiness check complete ===")
				return
			}
			t.Logf("✗ Ingress responding but not ready yet, status: %d (need 2xx or 3xx)", statusCode)
		} else {
			t.Logf("[Attempt %d] Request failed: %v", attemptNum, err)
		}

		t.Logf("Waiting 5 seconds before retry...")
		time.Sleep(5 * time.Second)
	}

	t.Logf("=== Ingress readiness check FAILED ===")
	t.Fatalf("Ingress not ready after %v (made %d attempts)", timeout, attemptNum)
}

func setupExternalMasterkeySecret(t *testing.T,
	opts *k8s.KubectlOptions) map[string]string {

	secretName := "my-external-masterkey"
	masterkeyValue := "externalkey1234567890abcdef12345"

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, opts)
	require.NoError(t, err)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: opts.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"masterkey": masterkeyValue,
		},
	}

	_, err = clientset.CoreV1().Secrets(opts.Namespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)
	require.NoError(t, err)

	t.Logf("Created external masterkey secret: %s in namespace: %s",
		secretName, opts.Namespace)

	return map[string]string{
		"zitadel.masterkeySecretName": secretName,
	}
}

func deployPostgres(t *testing.T, kubeOptions *k8s.KubectlOptions) {
	postgresHelmOptions := &helm.Options{
		KubectlOptions: kubeOptions,
		SetValues: map[string]string{
			"auth.postgresPassword":              "postgres",
			"primary.persistence.enabled":        "false",
			"image.repository":                   "bitnamilegacy/postgresql",
			"volumePermissions.image.repository": "bitnamilegacy/os-shell",
			"metrics.image.repository":           "bitnamilegacy/postgres-exporter",
			"fullnameOverride":                   "db-postgresql",
		},
	}

	helm.Install(t, postgresHelmOptions,
		"oci://registry-1.docker.io/bitnamicharts/postgresql", "db")

	k8s.WaitUntilPodAvailable(t, kubeOptions, "db-postgresql-0", 60,
		3*time.Second)
}

func assertMasterkeySecret(t *testing.T, kubeOptions *k8s.KubectlOptions,
	secretName string, expected masterkeyExpected) {

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	require.NoError(t, err)

	secret := getMasterkeySecret(t, clientset, kubeOptions.Namespace,
		secretName)

	if expected.secretCreated {
		require.NotNil(t, secret, "Secret %q should exist", secretName)

		masterkeyData, exists := secret.Data["masterkey"]
		require.True(t, exists, "Secret %q should contain 'masterkey' key",
			secretName)
		require.NotEmpty(t, masterkeyData,
			"Masterkey in Secret %q should not be empty", secretName)

		masterkeyValue := string(masterkeyData)

		if expected.masterkeyGenerated {
			require.Len(t, masterkeyValue, 32,
				"Generated masterkey should be 32 characters in Secret %q",
				secretName)
			require.Regexp(t, "^[A-Za-z0-9]{32}$", masterkeyValue,
				"Generated masterkey should be 32 alphanumeric "+
					"characters in Secret %q", secretName)
		} else if expected.masterkeyValue != "" {
			require.Equal(t, expected.masterkeyValue, masterkeyValue,
				"Masterkey should match provided value in Secret %q",
				secretName)
		}

		if expected.immutable {
			require.True(t, secret.Immutable != nil && *secret.Immutable,
				"Secret %q should be marked as immutable", secretName)
		}

		if !expected.isExternal {
			hookAnnotation, hasHook := secret.Annotations["helm.sh/hook"]
			require.True(t, hasHook,
				"Secret %q should have helm hook annotation", secretName)
			require.Equal(t, "pre-install", hookAnnotation,
				"Hook should be 'pre-install' only in Secret %q",
				secretName)
		}
	} else {
		require.Nil(t, secret, "Secret %q should NOT exist", secretName)
	}
}

func getMasterkeySecret(t *testing.T, clientset *kubernetes.Clientset,
	namespace, secretName string) *corev1.Secret {

	secret, err := clientset.CoreV1().Secrets(namespace).Get(
		context.Background(),
		secretName,
		metav1.GetOptions{},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		require.NoError(t, err, "unexpected error retrieving secret %q",
			secretName)
	}
	return secret
}

func sanitizeName(name string) string {
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else if c == '_' || c == ' ' || c == '-' {
			result += "-"
		}
	}
	return result
}

func TestMasterkeySecretLogic(t *testing.T) {
	suite.Run(t, new(MasterkeyTestSuite))
}
