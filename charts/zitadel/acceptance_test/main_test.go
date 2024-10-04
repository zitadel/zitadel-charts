package acceptance_test

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	terratesting "github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var ChartPath string

func TestMain(m *testing.M) {
	t := &mockT{*log.New(os.Stderr, "", 0)}
	var err error
	ChartPath, err = filepath.Abs("..")
	require.NoError(t, err)
	helm.AddRepo(t, &helm.Options{}, Postgres.Name, Postgres.RepoUrl)
	helm.AddRepo(t, &helm.Options{}, Cockroach.Name, Cockroach.RepoUrl)
	_, err = helm.RunHelmCommandAndGetOutputE(t, &helm.Options{}, "dependencies", "build", ChartPath)
	require.NoError(t, err)
	m.Run()
}

var _ terratesting.TestingT = &mockT{}

type mockT struct {
	log.Logger
}

func (m *mockT) Fail() {
	m.Logger.Fatal("Fail() called")
}

func (m *mockT) FailNow() {
	m.Logger.Fatal("FailNow() called")
}

func (m *mockT) Error(args ...interface{}) {
	m.Logger.Fatalf("FailNow(%v) called", args)
}

func (m *mockT) Errorf(format string, args ...interface{}) {
	m.Logger.Fatalf("FailNow(%s, %v) called", format, args)
}

func (m *mockT) Name() string {
	return "TestMain"
}
