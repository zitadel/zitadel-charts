package integration

import (
	"context"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	"sync"
	"time"
)

func (s *integrationTest) awaitAvailability(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.context, 5*time.Minute)
	defer cancel()

	wg := sync.WaitGroup{}
	for _, p := range pods {
		wg.Add(1)
		go func(pod corev1.Pod) {
			k8s.WaitUntilPodAvailable(s.T(), s.kubeOptions, pod.Name, 180, time.Second)
			wg.Done()
		}(p)
	}
	wait(ctx, s.T(), &wg, "availability")
}
