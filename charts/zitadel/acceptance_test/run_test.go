package acceptance_test

import (
	"context"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestZitadelInstallation() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	t := suite.T()

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
