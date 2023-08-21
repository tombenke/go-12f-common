package env

import (
	"os"
	"strconv"

	"github.com/tombenke/go-12f-common/log"
)

var logger = log.Logger

// GetEnvWithDefault gets the value of the `envVarName` environment variable and return with it.
// If there is no such variable defined in the environment, then return with the `defaultValue`.
func GetEnvWithDefault(envVarName string, defaultValue string) string {
	value, ok := os.LookupEnv(envVarName)
	if !ok {
		value = defaultValue
	}
	return value
}

// GetEnvWithDefaultUint gets the value of the `envVarName` environment variable and return with it as an `uint` type value.
// If there is no such variable defined in the environment, then return with the `defaultValue`.
func GetEnvWithDefaultUint(envVarName string, defaultValueStr string) uint64 {
	strValue := GetEnvWithDefault(envVarName, defaultValueStr)
	value, err := strconv.ParseUint(strValue, 10, 32)
	if err != nil {
		logger.WithError(err).Errorf("conversion error of %s", strValue)
	}
	return value
}
