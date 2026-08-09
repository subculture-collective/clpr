package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackgroundJobRunnerRejectsWorkAtCapacity(t *testing.T) {
	runner := newBackgroundJobRunner(0, 1)
	require.True(t, runner.Submit(func() {}))
	require.False(t, runner.Submit(func() {}))
}
