package acceptance_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
)

type hookFunc func(*ConfigurationTest)

type ConfigurationTest struct {
	suite.Suite
	Ctx                                       context.Context
	log                                       *logger.Logger
	KubeOptions                               *k8s.KubectlOptions
	KubeClient                                *kubernetes.Clientset
	Scheme, Domain                            string
	Port                                      uint16
	zitadelValues                             []string
	dbChart                                   *databaseChart
	zitadelRelease, dbRelease                 string
	beforeFunc, afterDBFunc, afterZITADELFunc hookFunc
}

func (c *ConfigurationTest) APIBaseURL() string {
	return fmt.Sprintf(`%s://%s:%d`, c.Scheme, c.Domain, c.Port)
}

type databaseChart struct {
	valuesFile, RepoUrl, Name, version string
	testValues                         map[string]string
}

var (
	Cockroach = databaseChart{
		RepoUrl: "https://charts.cockroachdb.com/",
		Name:    "cockroachdb",
		version: "13.0.1",
		testValues: map[string]string{
			"statefulset.replicas": "1",
			"conf.single-node":     "true",
		},
	}
	Postgres = databaseChart{
		RepoUrl: "https://charts.bitnami.com/bitnami",
		Name:    "postgresql",
		version: "12.10.0",
	}
)

func (d databaseChart) WithValues(valuesFile string) *databaseChart {
	d.valuesFile = valuesFile
	return &d
}

func Configure(
	t *testing.T,
	namespace string,
	dbChart *databaseChart,
	zitadelValues []string,
	externalDomain string,
	externalPort uint16,
	externalSecure bool,
	before, afterDB, afterZITADEL hookFunc,
) *ConfigurationTest {
	kubeOptions := k8s.NewKubectlOptions("", "", namespace)
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	if err != nil {
		t.Fatal(err)
	}
	externalScheme := "http"
	if externalSecure {
		externalScheme = "https"
	}
	cfg := &ConfigurationTest{
		Ctx:              context.Background(),
		log:              logger.New(logger.Terratest),
		KubeOptions:      kubeOptions,
		KubeClient:       clientset,
		zitadelValues:    zitadelValues,
		zitadelRelease:   "zitadel-test",
		dbChart:          dbChart,
		dbRelease:        "db",
		beforeFunc:       before,
		afterDBFunc:      afterDB,
		afterZITADELFunc: afterZITADEL,
		Domain:           externalDomain,
		Port:             externalPort,
		Scheme:           externalScheme,
	}
	cfg.SetT(t)
	return cfg
}
