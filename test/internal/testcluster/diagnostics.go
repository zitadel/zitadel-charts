package testcluster

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

const (
	// captureTimeout bounds how long collecting diagnostics may take, so a
	// struggling API server cannot hold up the rest of the test run.
	captureTimeout = 2 * time.Minute

	// logTailLines is the number of trailing lines kept per container log.
	logTailLines int64 = 5000
)

// CaptureNamespace writes the state of the test namespace into dir so it
// survives the runner, whose k3s cluster is destroyed after the job. It collects
// the pods and their status, the logs of every container including init
// containers, the logs of the previous instance of restarted containers, the
// namespace events, and the cluster nodes with their conditions. Failures to
// collect are logged and never fail the test, because the test has already
// failed by the time this runs.
func CaptureNamespace(t *testing.T, k *k8s.KubectlOptions, dir string) {
	t.Helper()

	client, err := k8s.GetKubernetesClientFromOptionsE(t, k)
	if err != nil {
		t.Logf("Skipping cluster state capture, cannot create a Kubernetes client: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), captureTimeout)
	defer cancel()

	if err := captureNamespace(ctx, client, k.Namespace, dir); err != nil {
		t.Logf("Cluster state capture for namespace %s was incomplete: %v", k.Namespace, err)
	}
	t.Logf("Cluster state for namespace %s written to %s", k.Namespace, dir)
}

func captureNamespace(ctx context.Context, client kubernetes.Interface, namespace, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	var errs []error

	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		errs = append(errs, fmt.Errorf("list pods: %w", err))
	} else {
		for i := range pods.Items {
			pods.Items[i].ManagedFields = nil
		}
		errs = append(errs, writeYAML(filepath.Join(dir, "pods.yaml"), pods))
		for i := range pods.Items {
			errs = append(errs, capturePodLogs(ctx, client, &pods.Items[i], dir)...)
		}
	}

	errs = append(errs, captureEvents(ctx, client, namespace, dir))
	errs = append(errs, captureNodes(ctx, client, dir))

	return errors.Join(errs...)
}

func capturePodLogs(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod, dir string) []error {
	restarts := make(map[string]int32)
	for _, status := range pod.Status.InitContainerStatuses {
		restarts[status.Name] = status.RestartCount
	}
	for _, status := range pod.Status.ContainerStatuses {
		restarts[status.Name] = status.RestartCount
	}

	var containers []string
	for _, c := range pod.Spec.InitContainers {
		containers = append(containers, c.Name)
	}
	for _, c := range pod.Spec.Containers {
		containers = append(containers, c.Name)
	}

	var errs []error
	for _, name := range containers {
		file := filepath.Join(dir, fmt.Sprintf("%s_%s.log", pod.Name, name))
		errs = append(errs, writeLog(ctx, client, pod, name, false, file))

		if restarts[name] > 0 {
			previous := filepath.Join(dir, fmt.Sprintf("%s_%s.previous.log", pod.Name, name))
			errs = append(errs, writeLog(ctx, client, pod, name, true, previous))
		}
	}
	return errs
}

func writeLog(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod, container string, previous bool, file string) error {
	tail := logTailLines
	stream, err := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container:  container,
		Previous:   previous,
		Timestamps: true,
		TailLines:  &tail,
	}).Stream(ctx)
	if err != nil {
		return fmt.Errorf("logs of %s/%s (previous=%t): %w", pod.Name, container, previous, err)
	}
	defer func() { _ = stream.Close() }()

	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, stream); err != nil {
		return fmt.Errorf("read logs of %s/%s: %w", pod.Name, container, err)
	}
	return nil
}

func captureEvents(ctx context.Context, client kubernetes.Interface, namespace, dir string) error {
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}

	sort.SliceStable(events.Items, func(i, j int) bool {
		return eventTime(events.Items[i]).Before(eventTime(events.Items[j]))
	})

	var out strings.Builder
	for _, e := range events.Items {
		fmt.Fprintf(&out, "%s\t%s\t%s\t%s/%s\t%s\n",
			eventTime(e).Format(time.RFC3339), e.Type, e.Reason,
			strings.ToLower(e.InvolvedObject.Kind), e.InvolvedObject.Name, e.Message)
	}
	return os.WriteFile(filepath.Join(dir, "events.txt"), []byte(out.String()), 0o644)
}

func eventTime(e corev1.Event) time.Time {
	switch {
	case !e.LastTimestamp.IsZero():
		return e.LastTimestamp.Time
	case !e.EventTime.IsZero():
		return e.EventTime.Time
	default:
		return e.CreationTimestamp.Time
	}
}

func captureNodes(ctx context.Context, client kubernetes.Interface, dir string) error {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}
	for i := range nodes.Items {
		nodes.Items[i].ManagedFields = nil
	}
	return writeYAML(filepath.Join(dir, "nodes.yaml"), nodes)
}

func writeYAML(file string, obj any) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(file), err)
	}
	return os.WriteFile(file, data, 0o644)
}
