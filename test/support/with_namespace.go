package support

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zitadel/zitadel-charts/test/internal/testcluster"
)

// WithNamespace creates a unique ephemeral namespace for a test, initializes
// a Kubernetes client and logger, and passes the resulting Env to the callback.
// The namespace is cleaned up when the test finishes unless the test failed,
// in which case it is preserved for debugging.
func WithNamespace(t *testing.T, fn func(*Env)) {
	t.Helper()

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		client, err := k8s.GetKubernetesClientFromOptionsE(t, k)
		require.NoError(t, err)

		config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		require.NoError(t, err)
		dynClient, err := dynamic.NewForConfig(config)
		require.NoError(t, err)

		env := &Env{
			Ctx:           ctx,
			Namespace:     k.Namespace,
			Kube:          k,
			Client:        client,
			DynamicClient: dynClient,
			Logger:        logger.New(logger.Terratest),
		}
		fn(env)
	})
}

var helmNameRegex = regexp.MustCompile(`[^a-z0-9\-]`)

// MakeRelease generates a Helm-compatible release name by combining the base
// name, suffix, and a random unique identifier. The result is normalized to
// RFC 1123 DNS naming standards, limited to 53 characters maximum, and ensures
// proper formatting for Helm release naming conventions.
func (env *Env) MakeRelease(baseName, suffix string) string {
	const maxHelmNameLength = 53

	normalizeComponent := func(input string) string {
		if input == "" {
			return ""
		}

		normalized := strings.ToLower(strings.TrimSpace(input))
		normalized = strings.ReplaceAll(normalized, "_", "-")
		normalized = helmNameRegex.ReplaceAllString(normalized, "-")

		for strings.Contains(normalized, "--") {
			normalized = strings.ReplaceAll(normalized, "--", "-")
		}

		return strings.Trim(normalized, "-")
	}

	normalizedBase := normalizeComponent(baseName)
	normalizedSuffix := normalizeComponent(suffix)
	uniqueId := strings.ToLower(random.UniqueId())

	components := []string{normalizedBase, normalizedSuffix, uniqueId}
	releaseName := strings.Join(components, "-")

	if len(releaseName) > maxHelmNameLength {
		releaseName = releaseName[:maxHelmNameLength]
	}

	return strings.Trim(releaseName, "-")
}
