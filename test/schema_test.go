package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSchemaInSync verifies values.schema.json matches values.yaml.
//
// Regenerate: helm schema -f charts/zitadel/values.yaml -o charts/zitadel/values.schema.json --draft 2020 --use-helm-docs
func TestSchemaInSync(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller(0) failed; cannot determine test file path")
	chartPath := filepath.Join(filepath.Dir(file), "..", "charts", "zitadel")
	valuesFile := filepath.Join(chartPath, "values.yaml")
	schemaFile := filepath.Join(chartPath, "values.schema.json")

	committed, err := os.ReadFile(schemaFile)
	require.NoError(t, err)

	generatedFile := filepath.Join(t.TempDir(), "values.schema.json")
	output, err := exec.Command("helm", "schema", "-f", valuesFile, "-o", generatedFile, "--draft", "2020", "--use-helm-docs").CombinedOutput()
	require.NoError(t, err, string(output))

	generated, err := os.ReadFile(generatedFile)
	require.NoError(t, err)

	require.JSONEq(t, string(committed), string(generated), "run: helm schema -f charts/zitadel/values.yaml -o charts/zitadel/values.schema.json --draft 2020 --use-helm-docs")
}
