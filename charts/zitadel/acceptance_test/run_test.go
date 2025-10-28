package acceptance_test

import (
	"context"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
)

func (suite *IntegrationSuite) TestZitadelInstallation() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	t := suite.T()
	if !t.Run("install", func(t *testing.T) {
		helm.Install(t, &helm.Options{
			KubectlOptions: suite.KubeOptions,
			ValuesFiles:    suite.ZitadelValues,
			SetValues: map[string]string{
				"replicaCount":       "1",
				"login.replicaCount": "1",
				"pdb.enabled":        "true",
			},
		}, suite.ZitadelChartPath, suite.ZitadelRelease)
	}) {
		t.FailNow()
	}
	if !t.Run("service-ready", func(t *testing.T) {
		k8s.WaitUntilServiceAvailable(t, suite.KubeOptions, suite.ZitadelRelease, 60, 2*time.Second)
	}) {
		t.FailNow()
	}

	if !t.Run("login", func(t *testing.T) {
		suite.login(ctx, t)
	}) {
		t.FailNow()
	}

	if !t.Run("grpc", func(t *testing.T) {
		assertGRPCWorks(ctx, t, suite, "iam-admin")
	}) {
		t.FailNow()
	}
}
