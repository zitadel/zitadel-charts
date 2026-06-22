package acceptance_test

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	grpchelper "github.com/zitadel/zitadel-charts/test/acceptance/helpers/grpc"
	httphelper "github.com/zitadel/zitadel-charts/test/acceptance/helpers/http"
)

// CheckAuthenticatedAPI verifies that both HTTP and gRPC authenticated API
// endpoints work using the declarative admin-client system user introduced in
// chart v11. The chart registers admin-client as a SystemAPIUsers entry
// (IAM_OWNER) and stores its keypair in a kubernetes.io/tls secret. There is no
// imperatively-created machine key anymore.
//
// Unlike the legacy machine-user flow (which exchanged a service-account JWT
// for an access token at the token endpoint), a ZITADEL system user signs a JWT
// with its private key and uses that JWT directly as the bearer token. The JWT
// claims must be iss=sub=<system user name> and the audience must contain the
// external API URL (scheme + host + port).
//
// This check validates:
//   - The admin-service-key secret is provisioned with a usable private key
//   - System-user JWT assertion generation and RS256 signing
//   - Bearer authentication on both the HTTP and gRPC management APIs
//
// secretName is the tls secret holding the keypair (e.g.
// "<release>-admin-service-key"), keyField is the data key for the PEM private
// key ("tls.key"), and systemUserName is the SystemAPIUsers name ("admin-client").
func CheckAuthenticatedAPI(ctx context.Context, t *testing.T, k *k8s.KubectlOptions, apiBaseURL, secretName, keyField, systemUserName string) {
	t.Helper()

	secret := k8s.GetSecret(t, k, secretName)
	keyPEM := secret.Data[keyField]
	require.NotEmpty(t, keyPEM, "key %s in secret %s is empty", keyField, secretName)

	token, err := signSystemUserJWT(keyPEM, systemUserName, apiBaseURL)
	require.NoError(t, err, "failed to sign system-user JWT")

	authCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		httpErr := callAuthenticatedHTTP(authCtx, token, apiBaseURL)
		if !assert.NoError(collect, httpErr) {
			return
		}
		grpcErr := callAuthenticatedGRPC(authCtx, token, apiBaseURL)
		assert.NoError(collect, grpcErr)
	}, 1*time.Minute, time.Second, "calling authenticated endpoints failed for a minute")
}

// signSystemUserJWT builds and RS256-signs a ZITADEL system-user assertion. The
// JWT is used directly as the bearer token against the management API. The
// audience must be the external API base URL (including port), matching the
// client_id ZITADEL derives from the request host.
func signSystemUserJWT(keyPEM []byte, systemUserName, audience string) (string, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return "", fmt.Errorf("no PEM block found in private key")
	}
	var key *rsa.PrivateKey
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		key = k
	} else {
		parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return "", fmt.Errorf("parsing RSA private key: %w", err2)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("private key is not RSA")
		}
		key = rsaKey
	}

	b64 := func(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
	header, err := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims, err := json.Marshal(map[string]interface{}{
		"iss": systemUserName,
		"sub": systemUserName,
		"aud": audience,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	signingInput := b64(header) + "." + b64(claims)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}
	return signingInput + "." + b64(signature), nil
}

func callAuthenticatedHTTP(ctx context.Context, token, apiBaseURL string) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statusCode, _, err := httphelper.Get(checkCtx, fmt.Sprintf("%s/management/v1/languages", apiBaseURL),
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return fmt.Errorf("expected status 200 at authenticated endpoint, but got %d", statusCode)
	}
	return nil
}

func callAuthenticatedGRPC(ctx context.Context, token, apiBaseURL string) error {
	conn, err := grpchelper.Dial(ctx, apiBaseURL)
	if err != nil {
		return fmt.Errorf("couldn't create gRPC connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	authCtx := grpchelper.WithBearerToken(ctx, token)
	var resp emptypb.Empty
	err = grpchelper.Invoke(authCtx, conn, "/zitadel.management.v1.ManagementService/GetSupportedLanguages", &emptypb.Empty{}, &resp)
	if err != nil {
		return fmt.Errorf("authenticated gRPC call failed: %w", err)
	}
	return nil
}
