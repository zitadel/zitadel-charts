package acceptance_test

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
)

func (s *ConfigurationTest) awaitReadiness(t *testing.T, pods []corev1.Pod) {
	for _, p := range pods {
		t.Run("pod "+p.Name, func(t *testing.T) {
			k8s.WaitUntilPodAvailable(t, s.KubeOptions, p.Name, 300, time.Second)
		})
	}
}
