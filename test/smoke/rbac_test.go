package smoke_test_test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

func TestRBACLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":         support.DigestTag,
			"zitadel.masterkey": "01234567890123456789012345678901",
		},
	}

	releaseName := "rbac-labels"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/rbac_zitadel.yaml"})
	docs := strings.Split(rendered, "---")

	expectedLabels := support.ExpectedLabels(releaseName, "zitadel", support.ExpectedVersion, "", nil)

	foundRole := false
	foundRoleBinding := false

	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var typeMeta struct {
			Kind string `json:"kind"`
		}
		require.NoError(t, yaml.Unmarshal([]byte(doc), &typeMeta))

		switch typeMeta.Kind {
		case "Role":
			var role rbacv1.Role
			require.NoError(t, yaml.Unmarshal([]byte(doc), &role))
			support.AssertLabels(t, role.Labels, expectedLabels)
			foundRole = true
		case "RoleBinding":
			var roleBinding rbacv1.RoleBinding
			require.NoError(t, yaml.Unmarshal([]byte(doc), &roleBinding))
			support.AssertLabels(t, roleBinding.Labels, expectedLabels)
			foundRoleBinding = true
		}
	}

	require.True(t, foundRole, "expected Role to be rendered")
	require.True(t, foundRoleBinding, "expected RoleBinding to be rendered")
}
