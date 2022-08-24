package installation

import (
	"context"
	"sync"
	"testing"
)

func wait(ctx context.Context, t *testing.T, wg *sync.WaitGroup, waitFor string) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				t.Fatalf("awaiting %s failed: %s", waitFor, err)
			}
			return
		}
	}
}
