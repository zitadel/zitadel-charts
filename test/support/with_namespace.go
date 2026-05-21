package support

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	wenv "github.com/mridang/wilhelm/env"
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
		config, err := clientcmd.BuildConfigFromFlags("", k.ConfigPath)
		require.NoError(t, err)

		e, err := wenv.NewEnvWithContext(ctx, config, k.Namespace)
		require.NoError(t, err)

		fn(&Env{Env: e})
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

	var components []string
	for _, c := range []string{normalizedBase, normalizedSuffix, uniqueId} {
		if c != "" {
			components = append(components, c)
		}
	}
	releaseName := strings.Join(components, "-")

	if len(releaseName) > maxHelmNameLength {
		releaseName = releaseName[:maxHelmNameLength]
	}

	return strings.Trim(releaseName, "-")
}
