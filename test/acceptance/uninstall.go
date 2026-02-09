package acceptance_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckUninstall performs ZITADEL Helm release uninstallation and verifies that
// no unexpected ZITADEL resources remain in the namespace afterward. This
// validates that the chart's cleanup is correctly configured, ensuring GitOps-
// friendly deployments where Helm fully manages the resource lifecycle.
//
// The function uninstalls the ZITADEL Helm release, then queries the Kubernetes
// API for resources matching the release name prefix. Only ZITADEL resources are
// checked - other resources (like PostgreSQL) are ignored since they're managed
// by separate Helm releases.
//
// Known resources that remain after uninstall (by design):
//   - Machine user secrets: Annotated with helm.sh/resource-policy=keep so they
//     survive reinstalls (credentials shouldn't be regenerated on upgrade)
//   - Helm hook resources: The chart uses pre-install/pre-upgrade hooks with
//     hook-delete-policy=before-hook-creation, which only deletes hooks when
//     creating new ones (not on uninstall). This includes: masterkey secret,
//     secrets-yaml secret, config configmaps, init/setup jobs, serviceaccounts,
//     roles, and rolebindings.
//
// The whitelist parameter allows tests to specify additional expected resources
// (e.g., test-created secrets that aren't part of the chart).
func CheckUninstall(ctx context.Context, t *testing.T, k *k8s.KubectlOptions, whitelist []string) {
	t.Helper()

	options := &helm.Options{
		KubectlOptions: k,
	}
	helm.Delete(t, options, zitadelRelease, true)

	remaining := findZitadelResources(ctx, t, k)

	whitelistSet := make(map[string]bool)
	for _, item := range whitelist {
		whitelistSet[item] = true
	}

	// Machine user secrets are annotated with helm.sh/resource-policy=keep
	// so they intentionally survive uninstall (see job_setup.yaml)
	whitelistSet["Secret/zitadel-admin-sa"] = true

	// Helm hook resources with hook-delete-policy=before-hook-creation remain
	// after uninstall. These are deleted only when new hooks are created (upgrade).
	// This is the chart's current design - hooks persist for debugging purposes.
	hookResources := []string{
		"Secret/zitadel-test-masterkey",
		"Secret/zitadel-test-secrets-yaml",
		"ConfigMap/zitadel-test-config-yaml",
		"ConfigMap/zitadel-test-login-config-dotenv",
		"Job/zitadel-test-init",
		"Job/zitadel-test-setup",
		"ServiceAccount/zitadel-test",
		"ServiceAccount/zitadel-test-login",
		"Role/zitadel-test",
		"RoleBinding/zitadel-test",
	}
	for _, res := range hookResources {
		whitelistSet[res] = true
	}

	var unexpected []string
	for _, res := range remaining {
		if !whitelistSet[res] {
			unexpected = append(unexpected, res)
		}
	}

	if len(unexpected) > 0 {
		t.Logf("Unexpected ZITADEL resources remain after uninstall: %v", unexpected)
		t.Logf("These may indicate chart bugs - resources should be deleted by helm uninstall")
	}
	require.Empty(t, unexpected, "unexpected ZITADEL resources remain after uninstall: %v", unexpected)
}

// findZitadelResources returns all resources in the namespace that match the
// ZITADEL release name prefix. This filters out non-ZITADEL resources like
// PostgreSQL so we only validate the ZITADEL chart's cleanup behavior.
func findZitadelResources(ctx context.Context, t *testing.T, k *k8s.KubectlOptions) []string {
	t.Helper()

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, k)
	require.NoError(t, err)

	namespace := k.Namespace
	opts := metav1.ListOptions{}
	var resources []string

	// Helper to check if resource name matches ZITADEL release
	isZitadelResource := func(name string) bool {
		return strings.HasPrefix(name, zitadelRelease)
	}

	secrets, err := clientset.CoreV1().Secrets(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, s := range secrets.Items {
		if isZitadelResource(s.Name) {
			resources = append(resources, fmt.Sprintf("Secret/%s", s.Name))
		}
	}

	configmaps, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, cm := range configmaps.Items {
		if isZitadelResource(cm.Name) {
			resources = append(resources, fmt.Sprintf("ConfigMap/%s", cm.Name))
		}
	}

	services, err := clientset.CoreV1().Services(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, svc := range services.Items {
		if isZitadelResource(svc.Name) {
			resources = append(resources, fmt.Sprintf("Service/%s", svc.Name))
		}
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, d := range deployments.Items {
		if isZitadelResource(d.Name) {
			resources = append(resources, fmt.Sprintf("Deployment/%s", d.Name))
		}
	}

	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, ss := range statefulsets.Items {
		if isZitadelResource(ss.Name) {
			resources = append(resources, fmt.Sprintf("StatefulSet/%s", ss.Name))
		}
	}

	jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, j := range jobs.Items {
		if isZitadelResource(j.Name) {
			resources = append(resources, fmt.Sprintf("Job/%s", j.Name))
		}
	}

	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, pvc := range pvcs.Items {
		if isZitadelResource(pvc.Name) {
			resources = append(resources, fmt.Sprintf("PersistentVolumeClaim/%s", pvc.Name))
		}
	}

	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, ing := range ingresses.Items {
		if isZitadelResource(ing.Name) {
			resources = append(resources, fmt.Sprintf("Ingress/%s", ing.Name))
		}
	}

	serviceaccounts, err := clientset.CoreV1().ServiceAccounts(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, sa := range serviceaccounts.Items {
		if isZitadelResource(sa.Name) {
			resources = append(resources, fmt.Sprintf("ServiceAccount/%s", sa.Name))
		}
	}

	roles, err := clientset.RbacV1().Roles(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, r := range roles.Items {
		if isZitadelResource(r.Name) {
			resources = append(resources, fmt.Sprintf("Role/%s", r.Name))
		}
	}

	rolebindings, err := clientset.RbacV1().RoleBindings(namespace).List(ctx, opts)
	require.NoError(t, err)
	for _, rb := range rolebindings.Items {
		if isZitadelResource(rb.Name) {
			resources = append(resources, fmt.Sprintf("RoleBinding/%s", rb.Name))
		}
	}

	pdbs, err := clientset.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, opts)
	if err == nil {
		for _, pdb := range pdbs.Items {
			if isZitadelResource(pdb.Name) {
				resources = append(resources, fmt.Sprintf("PodDisruptionBudget/%s", pdb.Name))
			}
		}
	}

	return resources
}
