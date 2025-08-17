package env

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVarKey   string
		setEnvVar   bool
		fallbackVal string
		expectedVal string
	}{
		{"should fetch value from env", "TEST_GET_ENV_VAR", true, "fallback value", "some test value"},
		{"should fall to fallbackValue", "TEST_GET_ENV_VAR", false, "fallback value", "fallback value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnvVar {
				os.Setenv(tt.envVarKey, tt.expectedVal)
			}
			assert.Equal(t, tt.expectedVal, GetEnv(tt.envVarKey, tt.fallbackVal))
			if tt.setEnvVar {
				os.Unsetenv(tt.envVarKey)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expected int
	}{
		{"should parse to 10", "10", 10},
		{"should parse to 0", "0", 0},
		{"should parse to -10", "-10", -10},
		{"should overflow to 2^(63-1)", "18446744073709551616", 9223372036854775807},
		{"should overflow to -1 * 2^63", "-18446744073709551616", -9223372036854775808},
		{"should fallback to 0 for invalid value (float)", "561.54", 0},
		{"should fallback to 0 for invalid value (string)", "not-integer", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetEnvInt("DOES_NOT_MATTER", tt.val))
		})
	}
}
