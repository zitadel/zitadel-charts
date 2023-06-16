package installation

import (
	"context"
	"sync"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
)

func (s *ConfigurationTest) awaitReadiness(pods []corev1.Pod) {
	ctx, cancel := context.WithTimeout(s.Ctx, 5*time.Minute)
	defer cancel()
	wg := sync.WaitGroup{}
	for _, p := range pods {
		wg.Add(1)
		go func(pod corev1.Pod) {
			k8s.WaitUntilPodAvailable(s.T(), s.KubeOptions, pod.Name, 300, time.Second)
			wg.Done()
		}(p)
	}
	wait(ctx, s.T(), &wg, "readiness")
}
