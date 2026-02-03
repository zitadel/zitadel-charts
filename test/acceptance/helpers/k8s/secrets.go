package k8s

import (
	"context"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateTLSSecret creates a Kubernetes TLS secret with the provided certificate
// and key data. The secret includes the CA certificate under ca.crt for chain
// verification. The secret is created in the namespace from KubectlOptions.
func CreateTLSSecret(t *testing.T, k *k8s.KubectlOptions, name string, caCert, tlsCert, tlsKey []byte) {
	t.Helper()
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, k)
	require.NoError(t, err, "failed to create k8s client")

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: k.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"ca.crt":  caCert,
			"tls.crt": tlsCert,
			"tls.key": tlsKey,
		},
	}

	_, err = clientset.CoreV1().Secrets(k.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	require.NoError(t, err, "failed to create TLS secret %s", name)
}

// CreateOpaqueSecret creates a Kubernetes opaque secret with the provided
// string data. The secret is created in the namespace from KubectlOptions.
func CreateOpaqueSecret(t *testing.T, k *k8s.KubectlOptions, name string, data map[string]string) {
	t.Helper()
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, k)
	require.NoError(t, err, "failed to create k8s client")

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: k.Namespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: data,
	}

	_, err = clientset.CoreV1().Secrets(k.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	require.NoError(t, err, "failed to create opaque secret %s", name)
}
