package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDocsSync verifies that README.md is in sync with values.yaml by running
// helm-docs and comparing the output. If this test fails, run:
//
//	helm-docs --chart-search-root=charts
//
// Then commit the updated README.md.
func TestDocsSync(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to determine caller info")

	repoRoot := filepath.Join(filepath.Dir(filename), "..")
	chartPath := filepath.Join(repoRoot, "charts", "zitadel")
	readmePath := filepath.Join(chartPath, "README.md")

	originalContent, err := os.ReadFile(readmePath)
	require.NoError(t, err, "failed to read README.md")

	cmd := exec.Command("helm-docs", "--chart-search-root=charts")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "helm-docs failed: %s", string(output))

	regeneratedContent, err := os.ReadFile(readmePath)
	require.NoError(t, err, "failed to read regenerated README.md")

	err = os.WriteFile(readmePath, originalContent, 0644)
	require.NoError(t, err, "failed to restore original README.md")

	require.Equal(t, string(originalContent), string(regeneratedContent),
		"README.md is out of sync with values.yaml. To fix, run 'helm-docs --chart-search-root=charts' and commit the updated README.md.")
}
