package acceptance

import (
	"context"
	"sync"
	"testing"
	"time"
)

func Await(ctx context.Context, t *testing.T, wg *sync.WaitGroup, tries int, cb func(ctx context.Context) error) {
	err := cb(ctx)
	if err == nil {
		if wg != nil {
			wg.Done()
		}
		return
	}
	if tries == 0 {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	Await(ctx, t, wg, tries-1, cb)
}
