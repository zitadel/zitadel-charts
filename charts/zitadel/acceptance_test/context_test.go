package acceptance_test

import (
	"context"
	"fmt"
	"testing"
)

func cancelTest(ctx context.Context, cancel context.CancelCauseFunc, t *testing.T) {
	t.Helper()
	if t.Failed() {
		cancel(fmt.Errorf("test %s failed", t.Name()))
	}
	cancel(nil)
	if err := context.Cause(ctx); err != nil && err != context.Canceled {
		t.Errorf("context cancelled: %v", err)
	}
}
