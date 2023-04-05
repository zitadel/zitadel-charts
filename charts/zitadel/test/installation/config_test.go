//go:build integration
// +build integration

package installation_test

import (
	"context"
	"github.com/zitadel/zitadel-charts/charts/zitadel/test/installation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"testing"
)

func TestWithInlineSecrets(t *testing.T) {

	installation.TestConfiguration(t, nil, map[string]string{
		"zitadel.masterkey": "x123456789012345678901234567891y",
		"zitadel.secretConfig.Database.cockroach.User.Password": "xy",
		"zitadel.configmapConfig.ExternalPort":                  "8080",
		"zitadel.configmapConfig.ExternalSecure":                "false",
		"zitadel.configmapConfig.TLS.Enabled":                   "false",
		"pdb.enabled":                                           "true",
		"ingress.enabled":                                       "true",
	})
}

func TestWithReferencedSecrets(t *testing.T) {

	masterKeySecretName := "existing-zitadel-masterkey"
	masterKeySecretKey := "masterkey"
	zitadelConfigSecretName := "existing-zitadel-secrets"
	zitadelConfigSecretKey := "config-yaml"

	installation.TestConfiguration(t, func(ctx context.Context, namespace string, k8sClient *kubernetes.Clientset) error {

		if err := createSecret(ctx, namespace, k8sClient, masterKeySecretName, masterKeySecretKey, "x123456789012345678901234567891y"); err != nil {
			return err
		}

		return createSecret(ctx, namespace, k8sClient, zitadelConfigSecretName, zitadelConfigSecretKey, "ExternalSecure: false\n")

	}, map[string]string{
		"zitadel.masterkeySecretName":                           masterKeySecretName,
		"zitadel.secretConfig.Database.cockroach.User.Password": "xy",
		"zitadel.configSecretName":                              zitadelConfigSecretName,
		"zitadel.configmapConfig.ExternalPort":                  "8080",
		"zitadel.configmapConfig.TLS.Enabled":                   "false",
		"pdb.enabled":                                           "true",
		"ingress.enabled":                                       "true",
	})
}

func createSecret(ctx context.Context, namespace string, k8sClient *kubernetes.Clientset, name, key, value string) error {
	_, err := k8sClient.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		StringData: map[string]string{key: value},
	}, metav1.CreateOptions{})
	return err
}
