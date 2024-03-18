package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssertf(t *testing.T) {
	require.PanicsWithError(t, "assertf_test.go(11): 2 + 2 == 4", func() {
		Assertf(2+2 == 5, "2 + 2 == 4")
	})
}
