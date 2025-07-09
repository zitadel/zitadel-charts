package acceptance_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Await(ctx context.Context, t *testing.T, waitFor time.Duration, cb func(ctx context.Context) error) {
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		assert.NoError(collect, cb(ctx))
	}, waitFor, time.Second)
}
