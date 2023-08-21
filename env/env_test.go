package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/env"
)

func TestGetEnvWithDefault(t *testing.T) {

	defaultValue := "TEST_MISSING_STR_ENV_VAR value"
	assert.Equal(t, defaultValue, env.GetEnvWithDefault("TEST_MISSING_STR_ENV_VAR", defaultValue))
	assert.Equal(t, "TEST_STR_ENV_VAR value", env.GetEnvWithDefault("TEST_STR_ENV_VAR", ""))
}

func TestGetEnvWithDefaultUint(t *testing.T) {

	// Test with valid default value and valid env var
	assert.Equal(t, uint64(42), env.GetEnvWithDefaultUint("TEST_UINT_ENV_VAR", "0"))

	// Test with valid default value and missing env var
	defaultValue := uint64(24)
	assert.Equal(t, defaultValue, env.GetEnvWithDefaultUint("TEST_MISSING_UINT_ENV_VAR", "24"))

	// Test with invalid default value and missing env var
	assert.Equal(t, uint64(0), env.GetEnvWithDefaultUint("TEST_MISSING_UINT_ENV_VAR", "xxx"))
}
