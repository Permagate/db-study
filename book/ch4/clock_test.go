package ch4

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClockMap(t *testing.T) {
	clock := NewClockMap(3)

	require.Equal(t, 0, clock.Len())
	require.Equal(t, 3, clock.Cap())
}
