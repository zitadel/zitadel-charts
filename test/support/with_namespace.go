package support

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
)

// WithNamespace creates a unique ephemeral namespace for a test using the
// provided shared cluster connection. The namespace is automatically deleted
// when the test finishes via t.Cleanup. A logger is injected into the Env
// for consistent test output formatting.
func WithNamespace(testing *testing.T, cluster *Cluster, fn func(*Env)) {
	testing.Helper()

	namespace := "e2e-" + strings.ToLower(random.UniqueId())
	kubectlOptions := k8s.NewKubectlOptions(cluster.ConfigPath, cluster.ContextName, namespace)

	k8s.CreateNamespace(testing, kubectlOptions, namespace)
	testing.Cleanup(func() {
		_ = k8s.DeleteNamespaceE(testing, kubectlOptions, namespace)
	})

	env := &Env{
		Namespace: namespace,
		Kube:      kubectlOptions,
		Client:    cluster.Client,
		Logger:    logger.New(logger.Terratest),
	}
	fn(env)
}

var (
	helmNameRegex = regexp.MustCompile(`[^a-z0-9\-]`)
)

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
