package support

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// DigestTag is a sample image tag containing a digest, used to verify that
	// label rendering strips the digest portion. Use a real digest so Helm
	// installs can actually pull the image during smoke tests.
	DigestTag = "v4.2.0@sha256:4582be1a9eeae5823aad17f58a58746696c43bcc09851364f0028077ebcadadf"
	// ExpectedVersion is the tag value that should appear in app.kubernetes.io/version.
	ExpectedVersion = "v4.2.0"
)

// ExpectedLabels builds the standard label set used by the chart helpers,
// plus an optional component label and any extra overrides provided.
func ExpectedLabels(releaseName, appName, version, component string, extra map[string]string) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":       appName,
		"app.kubernetes.io/instance":   releaseName,
		"app.kubernetes.io/version":    version,
		"app.kubernetes.io/managed-by": "Helm",
	}

	if component != "" {
		labels["app.kubernetes.io/component"] = component
	}

	for k, v := range extra {
		labels[k] = v
	}

	return labels
}

// AssertLabels verifies the provided labels map contains all expected key/value
// pairs. It does not require exact equality, only that the expected subset
// exists.
func AssertLabels(t *testing.T, actual map[string]string, expected map[string]string) {
	t.Helper()

	require.NotNil(t, actual, "expected labels to be set")
	for key, expectedValue := range expected {
		require.Contains(t, actual, key, "label %s missing", key)
		require.Equal(t, expectedValue, actual[key], "label %s mismatch", key)
	}
}
