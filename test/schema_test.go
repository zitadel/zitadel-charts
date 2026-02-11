package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaInSync verifies values.schema.json matches values.yaml.
//
// Regenerate: helm schema -f charts/zitadel/values.yaml -o charts/zitadel/values.schema.json --draft 2020 --use-helm-docs --k8s-schema-version v1.30.0
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
	cmd := exec.Command("helm", "schema",
		"-f", valuesFile,
		"-o", generatedFile,
		"--draft", "2020",
		"--use-helm-docs",
		"--k8s-schema-version", "v1.30.0")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	generated, err := os.ReadFile(generatedFile)
	require.NoError(t, err)

	require.JSONEq(t, string(committed), string(generated),
		"run: helm schema -f charts/zitadel/values.yaml -o charts/zitadel/values.schema.json --draft 2020 --use-helm-docs --k8s-schema-version v1.30.0")
}

// TestSchemaFullyTyped ensures all fields in the schema have proper types.
// This prevents generic "type": "object" or "type": "array" definitions that
// lack structure. Primitive types and arrays of primitives are allowed, as
// are map[string]string types identified by their description annotation.
// Complex types must either define nested properties or reference external
// Kubernetes schemas via $ref. Some paths like configmapConfig are ignored
// as they are intentionally free-form.
func TestSchemaFullyTyped(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller(0) failed; cannot determine test file path")

	schemaFile := filepath.Join(filepath.Dir(file), "..", "charts", "zitadel", "values.schema.json")
	data, err := os.ReadFile(schemaFile)
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(data, &schema))

	ignored := map[string]bool{
		"/zitadel/configmapConfig":                   true,
		"/zitadel/secretConfig":                      true,
		"/extraManifests":                            true,
		"/metrics/serviceMonitor/tlsConfig":          true,
		"/metrics/serviceMonitor/relabellings":       true,
		"/metrics/serviceMonitor/metricRelabellings": true,
	}

	properties := schema["properties"].(map[string]any)

	var check func(props map[string]any, path string) []string
	check = func(props map[string]any, path string) []string {
		var failures []string

		for name, prop := range props {
			propPath := path + "/" + name
			if ignored[propPath] {
				continue
			}

			propMap, ok := prop.(map[string]any)
			if !ok {
				continue
			}

			if _, hasRef := propMap["$ref"]; hasRef {
				continue
			}

			propType, _ := propMap["type"].(string)
			desc, _ := propMap["description"].(string)

			switch propType {
			case "string", "integer", "boolean", "number", "null":
				continue

			case "array":
				items, hasItems := propMap["items"].(map[string]any)
				if !hasItems {
					failures = append(failures, propPath)
					continue
				}
				if _, hasRef := items["$ref"]; hasRef {
					continue
				}
				itemType, _ := items["type"].(string)
				switch itemType {
				case "string", "integer", "boolean", "number":
					continue
				case "object":
					itemDesc, _ := items["description"].(string)
					if strings.Contains(itemDesc, "(map[string]string)") {
						continue
					}
					if nested, ok := items["properties"].(map[string]any); ok {
						failures = append(failures, check(nested, propPath+"[]")...)
						continue
					}
					failures = append(failures, propPath)
				default:
					failures = append(failures, propPath)
				}

			case "object":
				if strings.Contains(desc, "(map[string]string)") {
					continue
				}
				if nested, ok := propMap["properties"].(map[string]any); ok {
					failures = append(failures, check(nested, propPath)...)
					continue
				}
				failures = append(failures, propPath)

			case "":
				if nested, ok := propMap["properties"].(map[string]any); ok {
					failures = append(failures, check(nested, propPath)...)
				}
			}
		}

		return failures
	}

	for _, path := range check(properties, "") {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			assert.Fail(t, "untyped field - add @schema $ref or itemRef annotation")
		})
	}
}
