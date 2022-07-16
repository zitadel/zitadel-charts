package integration

import (
	"bufio"
	"context"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
	"sync"
	"testing"
	"time"
)

func (s *integrationTest) awaitListening(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.context, 1*time.Minute)
	defer cancel()

	k8sClient, err := k8s.GetKubernetesClientE(s.T())
	if err != nil {
		s.T().Fatal(err)
	}

	podsClient := k8sClient.CoreV1().Pods(s.namespace)

	wg := sync.WaitGroup{}
	for _, pod := range pods {
		wg.Add(1)
		go awaitisteningMessage(s.context, s.T(), &wg, podsClient, pod)
	}
	wait(ctx, s.T(), &wg, "listening")
}

func awaitisteningMessage(ctx context.Context, t *testing.T, wg *sync.WaitGroup, client v1.PodInterface, pod corev1.Pod) {
	reader, err := client.GetLogs(pod.Name, &corev1.PodLogOptions{Follow: true}).Stream(ctx)
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logLine := scanner.Text()
		if strings.Contains(logLine, "server is listening") {
			wg.Done()
			return
		}
		if err = scanner.Err(); err != nil {
			t.Fatal(err)
		}
	}
}
