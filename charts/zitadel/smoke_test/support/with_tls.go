package support

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WithTLSSecret generates a self-signed X.509 certificate and creates a
// Kubernetes TLS secret in the test environment's namespace. The certificate
// is valid for 365 days and supports both DNS names and IP addresses as
// subject alternative names. The private key uses RSA 2048-bit encryption.
// This function is intended for testing TLS configurations without requiring
// external certificate authorities or pre-generated certificates.
func WithTLSSecret(testing *testing.T, env *Env, secretName string, hosts ...string) {
	testing.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(testing, err)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(testing, err)

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Zitadel Test"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(testing, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(testing, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: env.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": certPEM,
			"tls.key": keyPEM,
		},
	}

	_, err = env.Client.CoreV1().Secrets(env.Namespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)
	require.NoError(testing, err)
}
