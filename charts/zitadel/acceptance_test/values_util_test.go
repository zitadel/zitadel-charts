package acceptance_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func readValuesAsMap(t *testing.T, valuesFilePath string) map[string]string {
	t.Helper()

	valuesBytes, err := os.ReadFile(valuesFilePath)
	require.NoError(t, err, "failed to read values file: %s", valuesFilePath)

	var yamlData map[string]interface{}
	err = yaml.Unmarshal(valuesBytes, &yamlData)
	require.NoError(t, err, "failed to unmarshal YAML from: %s", valuesFilePath)

	result := make(map[string]string)
	flattenYAML("", yamlData, result)
	return result
}

func flattenYAML(prefix string, data interface{}, result map[string]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPrefix := key
			if prefix != "" {
				newPrefix = prefix + "." + key
			}
			flattenYAML(newPrefix, value, result)
		}
	case []interface{}:
		for i, item := range v {
			newPrefix := fmt.Sprintf("%s[%d]", prefix, i)
			flattenYAML(newPrefix, item, result)
		}
	case string:
		result[prefix] = v
	case int:
		result[prefix] = fmt.Sprintf("%d", v)
	case int64:
		result[prefix] = fmt.Sprintf("%d", v)
	case float64:
		result[prefix] = fmt.Sprintf("%g", v)
	case bool:
		result[prefix] = fmt.Sprintf("%t", v)
	case nil:
		result[prefix] = "null"
	default:
		result[prefix] = fmt.Sprintf("%v", v)
	}
}
