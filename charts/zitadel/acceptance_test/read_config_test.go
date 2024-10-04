package acceptance_test

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/random"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type Values struct {
	Zitadel struct {
		MasterkeySecretName string `yaml:"masterkeySecretName"`
		ConfigSecretName    string `yaml:"configSecretName"`
		ConfigmapConfig     struct {
			ExternalDomain string `yaml:"ExternalDomain"`
			ExternalPort   uint16 `yaml:"ExternalPort"`
			ExternalSecure bool   `yaml:"ExternalSecure"`
			FirstInstance  struct {
				Org struct {
					Machine struct {
						Machine struct {
							Username string `yaml:"Username"`
						} `yaml:"Machine"`
					} `yaml:"Machine"`
				} `yaml:"Org"`
			} `yaml:"FirstInstance"`
		} `yaml:"configmapConfig"`
	} `yaml:"zitadel"`
}

func readValues(t *testing.T, valuesFilePath string) (values Values) {
	// set default values like in the defaults.yaml
	values.Zitadel.ConfigmapConfig.ExternalDomain = "localhost"
	values.Zitadel.ConfigmapConfig.ExternalPort = 8080
	values.Zitadel.ConfigmapConfig.ExternalSecure = true
	valuesBytes, err := os.ReadFile(valuesFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(valuesBytes, &values); err != nil {
		t.Fatal(err)
	}
	return values
}

func newNamespaceIdentifier(testcase string) string {
	// if triggered by a github action the environment variable is set
	// we use it to better identify the test
	commitSHA, exist := os.LookupEnv("GITHUB_SHA")
	namespace := fmt.Sprintf("zitadel-test-%s-%s", testcase, strings.ToLower(random.UniqueId()))
	if exist {
		namespace += "-" + commitSHA
	}
	// max namespace length is 63 characters
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
	return truncateString(namespace, 63)
}

func truncateString(str string, num int) string {
	shortenStr := str
	if len(str) > num {
		shortenStr = str[0:num]
	}
	return shortenStr
}

func workingDirectory(exampleDir string) (workingDir, valuesFile string) {
	_, filename, _, _ := runtime.Caller(0)
	workingDir = filepath.Join(filename, "..", "..", "..", "..", "examples", exampleDir)
	valuesFile = filepath.Join(workingDir, "zitadel-values.yaml")
	return workingDir, valuesFile
}

func readConfig(t *testing.T, exampleDir string) (string, string, Values) {
	workingDir, valuesFile := workingDirectory(exampleDir)
	return workingDir, valuesFile, readValues(t, valuesFile)
}
