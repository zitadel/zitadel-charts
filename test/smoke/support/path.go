package support

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// ChartPath returns the absolute path to the Zitadel Helm chart.
func ChartPath(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to determine caller info for chart path resolution")
	}

	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	chartPath := filepath.Join(repoRoot, "charts", "zitadel")
	absPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)
	return absPath
}
