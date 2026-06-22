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

	"github.com/mridang/wilhelm/assert"
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
		Subject:      pkix.Name{CommonName: "test-service"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM
}

// masterkeyPresent asserts the masterkey secret holds a non-empty masterkey.
func masterkeyPresent() *assert.SecretAssertion {
	return &assert.SecretAssertion{
		Data: assert.Matching[map[string][]byte](
			gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
		),
	}
}

// tlsKeypairPresent asserts a kubernetes.io/tls secret holds a non-empty
// certificate and private key. Used for both the login-service-key and the
// admin-service-key, which the chart generates declaratively the same way.
func tlsKeypairPresent() *assert.SecretAssertion {
	return &assert.SecretAssertion{
		Data: assert.Matching[map[string][]byte](gomega.And(
			gomega.HaveKeyWithValue("tls.crt", gomega.Not(gomega.BeEmpty())),
			gomega.HaveKeyWithValue("tls.key", gomega.Not(gomega.BeEmpty())),
		)),
	}
}

// TestSecretsMatrix verifies the chart's declarative secret behaviour: the
// masterkey, the login-service-key, and (new in v11) the admin-service-key.
// As of v11 the chart no longer creates IAM machine-user secrets imperatively,
// so legacy FirstInstance.Org.Machine config must NOT produce any iam-admin /
// iam-admin-pat secret.
func TestSecretsMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		setValues        map[string]string
		preInstall       func(t *testing.T, env *support.Env)
		masterkey        *assert.SecretAssertion
		adminServiceKey  *assert.SecretAssertion
		loginServiceKey  *assert.SecretAssertion
		absentMachineKey string // secret name expected to be ABSENT (legacy machine key)
		absentMachinePat string // secret name expected to be ABSENT (legacy machine PAT)
	}{
		{
			// Out of the box: masterkey + a generated admin-service-key + a
			// generated login-service-key, all declarative secrets.
			name:            "default",
			setValues:       map[string]string{},
			masterkey:       masterkeyPresent(),
			adminServiceKey: tlsKeypairPresent(),
			loginServiceKey: tlsKeypairPresent(),
		},
		{
			// Backwards compatibility: an operator carrying the old
			// FirstInstance.Org.Machine block forward must NOT get an
			// imperatively-created iam-admin / iam-admin-pat secret anymore.
			// The declarative admin-service-key is what they get instead.
			name: "legacy-machine-config-ignored",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username":          "iam-admin",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":              "Admin Machine",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate": "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":           "1",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Pat.ExpirationDate":        "2029-01-01T00:00:00Z",
			},
			masterkey:        masterkeyPresent(),
			adminServiceKey:  tlsKeypairPresent(),
			loginServiceKey:  tlsKeypairPresent(),
			absentMachineKey: "iam-admin",
			absentMachinePat: "iam-admin-pat",
		},
		{
			// Disabling the admin key removes the generated secret entirely.
			name: "admin-key-disabled",
			setValues: map[string]string{
				"zitadel.adminServiceKey.enabled": "false",
			},
			masterkey:       masterkeyPresent(),
			adminServiceKey: nil, // generated admin-service-key must be absent
			loginServiceKey: tlsKeypairPresent(),
		},
		{
			// Bring-your-own admin key: when existingSecretName is set the chart
			// references it and generates nothing of its own.
			name: "admin-key-external-secret",
			setValues: map[string]string{
				"zitadel.adminServiceKey.existingSecretName": "my-admin-cert",
			},
			preInstall: func(t *testing.T, env *support.Env) {
				t.Helper()
				certPEM, keyPEM := generateSelfSignedTLS(t)
				_, err := env.Client.CoreV1().Secrets(env.Namespace).Create(
					env.Ctx,
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "my-admin-cert"},
						Type:       corev1.SecretTypeTLS,
						Data:       map[string][]byte{"tls.crt": certPEM, "tls.key": keyPEM},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			masterkey:       masterkeyPresent(),
			adminServiceKey: nil, // generated admin-service-key must be absent
			loginServiceKey: tlsKeypairPresent(),
		},
		{
			// Login disabled removes the login key; the admin key is unaffected.
			name: "login-disabled",
			setValues: map[string]string{
				"login.enabled": "false",
			},
			masterkey:       masterkeyPresent(),
			adminServiceKey: tlsKeypairPresent(),
			loginServiceKey: nil, // generated login-service-key must be absent
		},
		{
			// Bring-your-own login key: generated login-service-key must be absent.
			name: "login-external-secret",
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
						Data:       map[string][]byte{"tls.crt": certPEM, "tls.key": keyPEM},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			masterkey:       masterkeyPresent(),
			adminServiceKey: tlsKeypairPresent(),
			loginServiceKey: nil, // generated login-service-key must be absent
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

				if tc.adminServiceKey != nil {
					env.AssertPartial(t, releaseName+"-admin-service-key", *tc.adminServiceKey)
				} else {
					env.AssertNone(t, releaseName+"-admin-service-key", assert.SecretAssertion{})
				}

				if tc.loginServiceKey != nil {
					env.AssertPartial(t, releaseName+"-login-service-key", *tc.loginServiceKey)
				} else {
					env.AssertNone(t, releaseName+"-login-service-key", assert.SecretAssertion{})
				}

				if tc.absentMachineKey != "" {
					env.AssertNone(t, tc.absentMachineKey, assert.SecretAssertion{})
				}
				if tc.absentMachinePat != "" {
					env.AssertNone(t, tc.absentMachinePat, assert.SecretAssertion{})
				}
			})
		})
	}
}
