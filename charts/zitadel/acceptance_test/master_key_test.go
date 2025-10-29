package acceptance_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestExplicitMasterkeyValue(t *testing.T) {
	t.Parallel()
	namespace := newNamespaceIdentifier("explicit-masterkey-value")
	externalDomain := fmt.Sprintf("%s.127.0.0.1.sslip.io", namespace)

	expectedMasterkey := "abcd1234efgh5678ijkl9012mnop3456"

	setValues := map[string]string{
		"zitadel.masterkey":                      expectedMasterkey,
		"zitadel.configmapConfig.ExternalDomain": externalDomain,
		"ingress.enabled":                        "true",
		"ingress.className":                      "traefik",
		"login.ingress.enabled":                  "true",
		"login.ingress.className":                "traefik",
	}

	suite.Run(t, Configure(
		t,
		namespace,
		externalDomain,
		Postgres,
		"",
		setValues,
		nil,
		nil,
		func(suite *IntegrationSuite) {
			secretName := suite.ZitadelRelease + "-masterkey"
			assertMasterkeySecret(t, suite.KubeOptions, secretName, masterkeyExpected{
				secretCreated:      true,
				masterkeyValue:     expectedMasterkey,
				masterkeyGenerated: false,
				immutable:          true,
				isExternal:         false,
			})

			ctx := context.Background()
			suite.login(ctx, t)
			assertGRPCWorks(ctx, t, suite, "iam-admin")
		},
	))
}

func TestExternalSecretReference(t *testing.T) {
	t.Parallel()
	namespace := newNamespaceIdentifier("external-secret-reference")
	externalDomain := fmt.Sprintf("%s.127.0.0.1.sslip.io", namespace)

	secretName := "my-external-masterkey"
	masterkeyValue := "externalkey1234567890abcdef12345"

	setValues := map[string]string{
		"zitadel.masterkeySecretName":            secretName,
		"zitadel.configmapConfig.ExternalDomain": externalDomain,
		"ingress.enabled":                        "true",
		"ingress.className":                      "traefik",
		"login.ingress.enabled":                  "true",
		"login.ingress.className":                "traefik",
	}

	suite.Run(t, Configure(
		t,
		namespace,
		externalDomain,
		Postgres,
		"",
		setValues,
		func(suite *IntegrationSuite) {
			createExternalMasterkeySecret(t, suite, secretName, masterkeyValue)
		},
		nil,
		func(suite *IntegrationSuite) {
			assertMasterkeySecret(t, suite.KubeOptions, secretName, masterkeyExpected{
				secretCreated: true,
				isExternal:    true,
			})

			ctx := context.Background()
			suite.login(ctx, t)
			assertGRPCWorks(ctx, t, suite, "iam-admin")
		},
	))
}

func TestBothMasterkeyAndSecretNameShouldFail(t *testing.T) {
	t.Parallel()
	namespace := newNamespaceIdentifier("both-set-should-fail")
	externalDomain := fmt.Sprintf("%s.127.0.0.1.sslip.io", namespace)

	helmOptions := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespace),
		SetValues: map[string]string{
			"zitadel.masterkey":                      "abcd1234efgh5678ijkl9012mnop3456",
			"zitadel.masterkeySecretName":            "my-external-masterkey",
			"zitadel.configmapConfig.ExternalDomain": externalDomain,
			"zitadel.configmapConfig.ExternalPort":   "443",
			"zitadel.configmapConfig.ExternalSecure": "true",
		},
	}

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	k8s.CreateNamespace(t, helmOptions.KubectlOptions, namespace)
	defer k8s.DeleteNamespace(t, helmOptions.KubectlOptions, namespace)

	err = helm.InstallE(t, helmOptions, chartPath, "zitadel-test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "set either .Values.zitadel.masterkey xor .Values.zitadel.masterkeySecretName")
}

type masterkeyExpected struct {
	secretCreated      bool
	masterkeyGenerated bool
	masterkeyValue     string
	immutable          bool
	isExternal         bool
}

func createExternalMasterkeySecret(t *testing.T, suite *IntegrationSuite, secretName, masterkeyValue string) {
	t.Helper()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: suite.KubeOptions.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"masterkey": masterkeyValue,
		},
	}

	_, err := suite.KubeClient.CoreV1().Secrets(suite.KubeOptions.Namespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)
	require.NoError(t, err)
}

func assertMasterkeySecret(t *testing.T, kubeOptions *k8s.KubectlOptions, secretName string, expected masterkeyExpected) {
	t.Helper()

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	require.NoError(t, err)

	secret := getMasterkeySecret(t, clientset, kubeOptions.Namespace, secretName)

	if expected.secretCreated {
		require.NotNil(t, secret, "Secret %q should exist", secretName)

		masterkeyData, exists := secret.Data["masterkey"]
		require.True(t, exists, "Secret %q should contain 'masterkey' key", secretName)
		require.NotEmpty(t, masterkeyData, "Masterkey in Secret %q should not be empty", secretName)

		masterkeyValue := string(masterkeyData)

		if expected.masterkeyGenerated {
			require.Len(t, masterkeyValue, 32, "Generated masterkey should be 32 characters in Secret %q", secretName)
			require.Regexp(t, "^[A-Za-z0-9]{32}$", masterkeyValue, "Generated masterkey should be 32 alphanumeric characters in Secret %q", secretName)
		} else if expected.masterkeyValue != "" {
			require.Equal(t, expected.masterkeyValue, masterkeyValue, "Masterkey should match provided value in Secret %q", secretName)
		}

		if expected.immutable {
			require.True(t, secret.Immutable != nil && *secret.Immutable, "Secret %q should be marked as immutable", secretName)
		}

		if !expected.isExternal {
			hookAnnotation, hasHook := secret.Annotations["helm.sh/hook"]
			require.True(t, hasHook, "Secret %q should have helm hook annotation", secretName)
			require.Equal(t, "pre-install", hookAnnotation, "Hook should be 'pre-install' only in Secret %q", secretName)
		}
	} else {
		require.Nil(t, secret, "Secret %q should NOT exist", secretName)
	}
}

func getMasterkeySecret(t *testing.T, clientset *kubernetes.Clientset, namespace, secretName string) *corev1.Secret {
	t.Helper()

	secret, err := clientset.CoreV1().Secrets(namespace).Get(
		context.Background(),
		secretName,
		metav1.GetOptions{},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		require.NoError(t, err, "unexpected error retrieving secret %q", secretName)
	}
	return secret
}
