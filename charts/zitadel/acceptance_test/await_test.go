package acceptance_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Awaitf(ctx context.Context, t *testing.T, waitFor time.Duration, cb func(ctx context.Context) error, msg string, args ...any) {
	require.EventuallyWithTf(t, func(collect *assert.CollectT) {
		if !assert.NoError(collect, cb(ctx)) {
			t.Logf("retrying in a second")
		}
	}, waitFor, time.Second, msg, args...)
}
