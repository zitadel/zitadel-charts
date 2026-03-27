package smoke_test_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

// generateSelfSignedTLS creates a PEM-encoded self-signed RSA certificate and
// private key suitable for use in a kubernetes.io/tls secret.
func generateSelfSignedTLS(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-login-service"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM
}

func TestSecretsMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		setValues       map[string]string
		preInstall      func(t *testing.T, env *support.Env)
		masterkey       *assert.SecretAssertion
		loginServiceKey *assert.SecretAssertion
		adminServiceKey *assert.SecretAssertion
	}{
		{
			name:      "default-all-enabled",
			setValues: map[string]string{},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
			adminServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
		},
		{
			name:      "minimal-no-setup",
			setValues: map[string]string{},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
			adminServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
		},
		{
			name: "x509-login-default",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
			adminServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
		},
		{
			name: "x509-disabled-when-login-disabled",
			setValues: map[string]string{
				"login.enabled": "false",
			},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			// loginServiceKey intentionally nil — should be absent when login is disabled
			adminServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
		},
		{
			name: "x509-external-login-secret",
			setValues: map[string]string{
				"login.enabled":                   "true",
				"login.loginServiceKeySecretName": "my-custom-cert",
			},
			preInstall: func(t *testing.T, env *support.Env) {
				t.Helper()
				certPEM, keyPEM := generateSelfSignedTLS(t)
				_, err := env.Client.CoreV1().Secrets(env.Namespace).Create(
					env.Ctx,
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "my-custom-cert"},
						Type:       corev1.SecretTypeTLS,
						Data: map[string][]byte{
							"tls.crt": certPEM,
							"tls.key": keyPEM,
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			// loginServiceKey intentionally nil — auto-generated secret should be absent
			adminServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
		},
		{
			name: "x509-external-admin-secret",
			setValues: map[string]string{
				"zitadel.adminServiceKeySecretName": "my-admin-cert",
			},
			preInstall: func(t *testing.T, env *support.Env) {
				t.Helper()
				certPEM, keyPEM := generateSelfSignedTLS(t)
				_, err := env.Client.CoreV1().Secrets(env.Namespace).Create(
					env.Ctx,
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "my-admin-cert"},
						Type:       corev1.SecretTypeTLS,
						Data: map[string][]byte{
							"tls.crt": certPEM,
							"tls.key": keyPEM,
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginServiceKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](gomega.And(
					gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
					gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
				)),
			},
			// adminServiceKey intentionally nil — auto-generated secret should be absent
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			support.WithNamespace(t, func(env *support.Env) {
				if tc.preInstall != nil {
					tc.preInstall(t, env)
				}

				releaseName := setup.InstallZitadel(t, env, tc.name, tc.setValues)

				if tc.masterkey != nil {
					env.AssertPartial(t, releaseName+"-masterkey", *tc.masterkey)
				}

				if tc.loginServiceKey != nil {
					env.AssertPartial(t, releaseName+"-login-service-key", *tc.loginServiceKey)
				} else {
					env.AssertNone(t, releaseName+"-login-service-key", assert.SecretAssertion{})
				}

				if tc.adminServiceKey != nil {
					env.AssertPartial(t, releaseName+"-admin-service-key", *tc.adminServiceKey)
				} else {
					env.AssertNone(t, releaseName+"-admin-service-key", assert.SecretAssertion{})
				}
			})
		})
	}
}
