package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/yaml"

	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
)

// TestSetupJobEnabledRendering verifies that setupJob.enabled gates the setup
// Job, mirroring initJob.enabled. When several Helm releases share one database,
// only the release that owns setup may render the setup hook, otherwise every
// release schedules its own setup Job against the same database on each
// install and upgrade. The checks are render-only to keep them fast and to
// avoid adding to the K3s cluster install budget.
func TestSetupJobEnabledRendering(t *testing.T) {
	t.Parallel()

	chartPath := setup.ChartPath(t)

	// minValues are the minimum values required for the chart to render the
	// job templates without errors.
	minValues := map[string]string{
		"zitadel.masterkey":                      "x123456789012345678901234567891y",
		"zitadel.configmapConfig.ExternalDomain": "auth.example.com",
		"zitadel.configmapConfig.ExternalPort":   "443",
		"zitadel.configmapConfig.TLS.Enabled":    "false",
	}

	// t.Run: enabled-by-default
	// Leaving the value unset must keep rendering the setup Job, so existing
	// installs keep running setup on every install and upgrade.
	t.Run("enabled-by-default", func(t *testing.T) {
		t.Parallel()

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: minValues},
			chartPath, "setup-job", []string{"templates/job_setup.yaml"})
		require.NoError(t, err)

		var job batchv1.Job
		require.NoError(t, yaml.Unmarshal([]byte(output), &job))
		require.Equal(t, "setup-job-zitadel-setup", job.Name)
	})

	// t.Run: disabled
	// With setupJob.enabled=false the setup Job must not be rendered at all.
	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		values := map[string]string{"setupJob.enabled": "false"}
		for key, value := range minValues {
			values[key] = value
		}

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "setup-job", nil)
		require.NoError(t, err)
		require.NotContains(t, output, "setup-job-zitadel-setup")
	})

	// t.Run: disabled-keeps-init-job
	// The two gates are independent: disabling setup must not drop the init
	// Job, and disabling init must not drop the setup Job.
	t.Run("disabled-keeps-init-job", func(t *testing.T) {
		t.Parallel()

		values := map[string]string{"setupJob.enabled": "false"}
		for key, value := range minValues {
			values[key] = value
		}

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "setup-job", []string{"templates/job_init.yaml"})
		require.NoError(t, err)
		require.Contains(t, output, "setup-job-zitadel-init")
	})

	// t.Run: init-disabled-keeps-setup-job
	t.Run("init-disabled-keeps-setup-job", func(t *testing.T) {
		t.Parallel()

		values := map[string]string{"initJob.enabled": "false"}
		for key, value := range minValues {
			values[key] = value
		}

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "setup-job", []string{"templates/job_setup.yaml"})
		require.NoError(t, err)
		require.Contains(t, output, "setup-job-zitadel-setup")
	})
}
