package smoke_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

func TestNginxConfiguration(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	testCases := []struct {
		name              string
		zitadelNginxImage string
		loginNginxImage   string
	}{
		{
			name:              "default-nginx-images",
			zitadelNginxImage: "",
			loginNginxImage:   "",
		},
		{
			name:              "custom-nginx-images",
			zitadelNginxImage: "nginx:1.26-alpine",
			loginNginxImage:   "nginx:1.25-alpine",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			support.WithNamespace(t, cluster, func(env *support.Env) {
				env.Logger.Logf(t, "namespace %q created; installing PostgreSQL…", env.Namespace)
				support.WithPostgres(t, env)

				releaseName := env.MakeRelease("zitadel-nginx", testCase.name)

				setValues := map[string]string{
					"zitadel.masterkey":                                         "x123456789012345678901234567891y",
					"zitadel.configmapConfig.ExternalDomain":                    "nginx-test.127.0.0.1.sslip.io",
					"zitadel.configmapConfig.ExternalPort":                      "443",
					"zitadel.configmapConfig.ExternalSecure":                    "true",
					"zitadel.configmapConfig.Database.Postgres.Host":            "db-postgresql",
					"zitadel.configmapConfig.Database.Postgres.Port":            "5432",
					"zitadel.configmapConfig.Database.Postgres.Database":        "zitadel",
					"zitadel.configmapConfig.Database.Postgres.MaxOpenConns":    "20",
					"zitadel.configmapConfig.Database.Postgres.MaxIdleConns":    "10",
					"zitadel.configmapConfig.Database.Postgres.MaxConnLifetime": "30m",
					"zitadel.configmapConfig.Database.Postgres.MaxConnIdleTime": "5m",
					"zitadel.configmapConfig.Database.Postgres.User.Username":   "postgres",
					"zitadel.configmapConfig.Database.Postgres.User.SSL.Mode":   "disable",
					"zitadel.configmapConfig.Database.Postgres.Admin.Username":  "postgres",
					"zitadel.configmapConfig.Database.Postgres.Admin.SSL.Mode":  "disable",
					"zitadel.selfSignedCert.enabled":                            "true",
					"login.selfSignedCert.enabled":                              "true",
					"ingress.enabled":                                           "false",
					"login.ingress.enabled":                                     "false",
					"service.appProtocol":                                       "kubernetes.io/https",
					"service.port":                                              "443",
				}

				zitadelImage := "nginx:1.27-alpine"
				loginImage := "nginx:1.27-alpine"

				if testCase.zitadelNginxImage != "" {
					setValues["zitadel.nginx.image"] = testCase.zitadelNginxImage
					zitadelImage = testCase.zitadelNginxImage
				}
				if testCase.loginNginxImage != "" {
					setValues["login.nginx.image"] = testCase.loginNginxImage
					loginImage = testCase.loginNginxImage
				}

				helmOptions := &helm.Options{
					KubectlOptions: env.Kube,
					SetValues:      setValues,
					ExtraArgs: map[string][]string{
						"upgrade": {"--install", "--wait", "--timeout", "30m", "--hide-notes"},
					},
				}

				require.NoError(t, helm.UpgradeE(t, helmOptions, chartPath, releaseName))

				assertContainer(t, env, releaseName, corev1.Container{
					Name:            "zitadel-reverse-proxy",
					Image:           zitadelImage,
					Command:         []string{"nginx", "-g", "daemon off;"},
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http2-server",
							ContainerPort: 443,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "nginx-cache",
							MountPath: "/var/cache/nginx",
						},
						{
							Name:      "nginx-run",
							MountPath: "/var/run",
						},
						{
							Name:      "tls-certs",
							ReadOnly:  true,
							MountPath: "/etc/nginx/certs",
						},
						{
							Name:      "nginx-conf",
							ReadOnly:  true,
							MountPath: "/etc/nginx/conf.d",
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/healthz",
								Port:   intstr.FromInt32(443),
								Scheme: corev1.URISchemeHTTPS,
							},
						},
						InitialDelaySeconds: 5,
						TimeoutSeconds:      1,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/healthz",
								Port:   intstr.FromInt32(443),
								Scheme: corev1.URISchemeHTTPS,
							},
						},
						InitialDelaySeconds: 5,
						TimeoutSeconds:      1,
						PeriodSeconds:       5,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					Resources:                corev1.ResourceRequirements{},
				})

				assertContainer(t, env, releaseName+"-login", corev1.Container{
					Name:            "zitadel-reverse-proxy",
					Image:           loginImage,
					Command:         []string{"nginx", "-g", "daemon off;"},
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http-server",
							ContainerPort: 443,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "nginx-cache",
							MountPath: "/var/cache/nginx",
						},
						{
							Name:      "nginx-run",
							MountPath: "/var/run",
						},
						{
							Name:      "tls-certs",
							ReadOnly:  true,
							MountPath: "/etc/nginx/certs",
						},
						{
							Name:      "nginx-conf",
							ReadOnly:  true,
							MountPath: "/etc/nginx/conf.d",
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/healthz",
								Port:   intstr.FromInt32(443),
								Scheme: corev1.URISchemeHTTPS,
							},
						},
						InitialDelaySeconds: 5,
						TimeoutSeconds:      1,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/healthz",
								Port:   intstr.FromInt32(443),
								Scheme: corev1.URISchemeHTTPS,
							},
						},
						InitialDelaySeconds: 5,
						TimeoutSeconds:      1,
						PeriodSeconds:       5,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					Resources:                corev1.ResourceRequirements{},
				})

				validateNginxLogs(t, env, releaseName)
				validateNginxLogs(t, env, releaseName+"-login")
			})
		})
	}
}

func assertContainer(t *testing.T, env *support.Env, deploymentName string, expected corev1.Container) {
	t.Helper()

	deployment, err := k8s.GetDeploymentE(t, env.Kube, deploymentName)
	require.NoError(t, err)

	var actual *corev1.Container
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == expected.Name {
			actual = &container
			break
		}
	}

	require.NotNil(t, actual, "container %s not found in deployment %s", expected.Name, deploymentName)
	require.Equal(t, expected, *actual, "container mismatch for %s in deployment %s", expected.Name, deploymentName)

	env.Logger.Logf(t, "✓ Verified container %s configuration for %s", expected.Name, deploymentName)
}

func validateNginxLogs(t *testing.T, env *support.Env, deploymentName string) {
	t.Helper()

	labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", deploymentName)

	podList, err := env.Client.CoreV1().Pods(env.Kube.Namespace).List(
		context.Background(),
		metav1.ListOptions{LabelSelector: labelSelector},
	)

	if err != nil {
		env.Logger.Logf(t, "failed to list pods for %s: %v", deploymentName, err)
		return
	}

	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, deploymentName) {
			validateContainerLogs(t, env, pod, "zitadel-reverse-proxy")
		}
	}
}

func validateContainerLogs(t *testing.T, env *support.Env, pod corev1.Pod, containerName string) {
	t.Helper()

	env.Logger.Logf(t, "==== Validating logs for pod=%s container=%s ====", pod.Name, containerName)

	logOutput, err := k8s.RunKubectlAndGetOutputE(
		t,
		env.Kube,
		"logs",
		pod.Name,
		"-c", containerName,
		"--tail=100",
	)

	if err != nil {
		env.Logger.Logf(t, "failed to get logs: %v", err)
		return
	}

	env.Logger.Logf(t, "%s", logOutput)

	lines := strings.Split(logOutput, "\n")
	var unexpectedWarnings []string

	for _, line := range lines {
		if strings.Contains(line, "[warn]") {
			if strings.Contains(line, `the "user" directive makes sense only if the master process runs with super-user privileges`) {
				continue
			}
			unexpectedWarnings = append(unexpectedWarnings, line)
		}
	}

	if len(unexpectedWarnings) > 0 {
		env.Logger.Logf(t, "❌ Found unexpected nginx warnings in pod=%s:", pod.Name)
		for _, warning := range unexpectedWarnings {
			env.Logger.Logf(t, "  %s", warning)
		}
		require.Fail(t, fmt.Sprintf("unexpected nginx warnings found in pod %s", pod.Name))
	} else {
		env.Logger.Logf(t, "✓ No unexpected nginx warnings found in pod=%s", pod.Name)
	}

	env.Logger.Logf(t, "==== End validation for pod=%s container=%s ====", pod.Name, containerName)
}
