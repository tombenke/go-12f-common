package log

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCaller(t *testing.T) {
	require.Equal(t, "github.com/tombenke/go-12f-common/log.TestGetCaller:10", testCaller())
}

func testCaller() string {
	return getCaller()
}
