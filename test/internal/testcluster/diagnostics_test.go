package testcluster

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

const testNamespace = "zitadel-test-abc"

func testPod(name string, restarts int32) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: testNamespace,
			ManagedFields: []metav1.ManagedFieldsEntry{{Manager: "kubectl"}},
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Name: "wait"}},
			Containers:     []corev1.Container{{Name: "zitadel"}},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{{Name: "wait"}},
			ContainerStatuses:     []corev1.ContainerStatus{{Name: "zitadel", RestartCount: restarts}},
		},
	}
}

func testEvent(name, reason, message string, at time.Time) *corev1.Event {
	return &corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Name: name, Namespace: testNamespace},
		Type:           corev1.EventTypeWarning,
		Reason:         reason,
		Message:        message,
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "zitadel-0"},
		LastTimestamp:  metav1.NewTime(at),
	}
}

// TestCaptureNamespace verifies that everything needed to explain a failed
// Login UI run is written to disk: the pod status, the logs of init and regular
// containers, the previous logs of a restarted container only, the events in
// time order, and the nodes. Objects from other namespaces must not leak in.
func TestCaptureNamespace(t *testing.T) {
	t.Parallel()

	now := time.Now()
	otherPod := testPod("other-0", 0)
	otherPod.Namespace = "some-other-namespace"

	client := fake.NewClientset(
		testPod("zitadel-0", 2),
		testPod("login-0", 0),
		otherPod,
		testEvent("late", "Unhealthy", "readiness probe failed", now),
		testEvent("early", "Pulled", "image pulled", now.Add(-time.Minute)),
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{
			Name:          "k3s-server",
			ManagedFields: []metav1.ManagedFieldsEntry{{Manager: "k3s"}},
		}},
	)

	dir := filepath.Join(t.TempDir(), "cluster")
	require.NoError(t, captureNamespace(context.Background(), client, testNamespace, dir))

	for _, name := range []string{
		"pods.yaml", "nodes.yaml", "events.txt",
		"zitadel-0_wait.log", "zitadel-0_zitadel.log", "zitadel-0_zitadel.previous.log",
		"login-0_wait.log", "login-0_zitadel.log",
	} {
		require.FileExists(t, filepath.Join(dir, name))
	}

	require.NoFileExists(t, filepath.Join(dir, "login-0_zitadel.previous.log"),
		"a container that never restarted has no previous instance to read")
	require.NoFileExists(t, filepath.Join(dir, "other-0_zitadel.log"),
		"pods of other namespaces must not be captured")

	for _, name := range []string{"pods.yaml", "nodes.yaml"} {
		content, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		require.NotContains(t, string(content), "managedFields", "server-side apply bookkeeping only adds noise")
	}

	logs, err := os.ReadFile(filepath.Join(dir, "zitadel-0_zitadel.log"))
	require.NoError(t, err)
	require.Equal(t, "fake logs", string(logs))

	events, err := os.ReadFile(filepath.Join(dir, "events.txt"))
	require.NoError(t, err)
	require.Contains(t, string(events), "readiness probe failed")
	require.Less(t, strings.Index(string(events), "image pulled"), strings.Index(string(events), "readiness probe failed"),
		"events must be listed oldest first")
}

// TestCaptureNamespaceKeepsGoingOnPartialFailure verifies that one failing API
// call does not stop the rest of the capture. A struggling cluster is exactly
// when the remaining evidence matters most.
func TestCaptureNamespaceKeepsGoingOnPartialFailure(t *testing.T) {
	t.Parallel()

	client := fake.NewClientset(
		testPod("zitadel-0", 0),
		testEvent("e", "Unhealthy", "readiness probe failed", time.Now()),
	)
	client.PrependReactor("list", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("api server unavailable")
	})

	dir := filepath.Join(t.TempDir(), "cluster")
	err := captureNamespace(context.Background(), client, testNamespace, dir)

	require.ErrorContains(t, err, "list pods")
	require.FileExists(t, filepath.Join(dir, "events.txt"))
	require.FileExists(t, filepath.Join(dir, "nodes.yaml"))
}
