package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetState() {
	viper.Reset()
	os.Unsetenv("RCMETERING_SERVER_HOST")
	os.Unsetenv("RCMETERING_SERVER_PORT")
	os.Unsetenv("RCMETERING_LOGGER_LEVEL")
	os.Unsetenv("RCMETERING_LOGGER_LOGFILE")
	os.Unsetenv("RCMETERING_LOGGER_MODE")
}

func createTestConfigFile(t *testing.T, content string) {
	err := os.WriteFile("config.env", []byte(content), 0644)
	require.NoError(t, err, "Failed to create test config file")
}

func removeTestConfigFile(t *testing.T) {
	err := os.Remove("config.env")
	if err != nil && !os.IsNotExist(err) {
		t.Logf("Failed to remove test config file: %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("Default Configuration", func(t *testing.T) {
		// Setup
		resetState()
		removeTestConfigFile(t)

		config, err := LoadConfig()

		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "localhost", config.Server.Host)
		assert.Equal(t, "8000", config.Server.Port)
		assert.Equal(t, INFO, config.Logger.Level)
		assert.Equal(t, "dev", config.Logger.Mode)
		assert.Equal(t, "rcmetering.log", config.Logger.LogFile) // Updated to match default
	})

	t.Run("Configuration from File", func(t *testing.T) {
		// Setup
		resetState()
		configContent := `
SERVER_HOST=testhost
SERVER_PORT=9000
LOGGER_LEVEL=warn
LOGGER_LOGFILE=/var/log/app.log
LOGGER_MODE=prod
`
		createTestConfigFile(t, configContent)
		defer removeTestConfigFile(t)

		// Test
		config, err := LoadConfig()

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "testhost", config.Server.Host)
		assert.Equal(t, "9000", config.Server.Port)
		assert.Equal(t, WARN, config.Logger.Level)
		assert.Equal(t, "/var/log/app.log", config.Logger.LogFile)
		assert.Equal(t, "prod", config.Logger.Mode)
	})

	t.Run("Environment Variables Override", func(t *testing.T) {
		// Setup
		resetState()
		configContent := `
SERVER_HOST=confighost
SERVER_PORT=8080
LOGGER_LEVEL=info
LOGGER_MODE=dev
`
		createTestConfigFile(t, configContent)
		defer removeTestConfigFile(t)

		// Set environment variables (with prefix)
		os.Setenv("RCMETERING_SERVER_HOST", "envhost")
		os.Setenv("RCMETERING_LOGGER_LEVEL", "error")
		defer resetState()

		// Test
		config, err := LoadConfig()

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "envhost", config.Server.Host) // Should use env var
		assert.Equal(t, "8080", config.Server.Port)    // Should use config file
		assert.Equal(t, ERROR, config.Logger.Level)    // Should use env var
		assert.Equal(t, "dev", config.Logger.Mode)     // Should use config file
	})
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name         string
		setupConfig  func() *Config
		expectError  bool
		errorMessage string
	}{
		{
			name: "Valid Configuration",
			setupConfig: func() *Config {
				return &Config{
					Server: ServerConfig{
						Host: "localhost",
						Port: "8000",
					},
					Logger: LoggerConfig{
						Level: INFO,
						Mode:  "dev",
					},
				}
			},
			expectError: false,
		},
		{
			name: "Missing Server Host",
			setupConfig: func() *Config {
				return &Config{
					Server: ServerConfig{
						Host: "",
						Port: "8000",
					},
					Logger: LoggerConfig{
						Level: INFO,
						Mode:  "dev",
					},
				}
			},
			expectError:  true,
			errorMessage: "missing server host",
		},
		{
			name: "Missing Server Port",
			setupConfig: func() *Config {
				return &Config{
					Server: ServerConfig{
						Host: "localhost",
						Port: "",
					},
					Logger: LoggerConfig{
						Level: INFO,
						Mode:  "dev",
					},
				}
			},
			expectError:  true,
			errorMessage: "missing server port",
		},
		{
			name: "Missing Logger Level",
			setupConfig: func() *Config {
				return &Config{
					Server: ServerConfig{
						Host: "localhost",
						Port: "8000",
					},
					Logger: LoggerConfig{
						Level: "",
						Mode:  "dev",
					},
				}
			},
			expectError:  true,
			errorMessage: "missing logger level",
		},
		{
			name: "Missing Logger Mode",
			setupConfig: func() *Config {
				return &Config{
					Server: ServerConfig{
						Host: "localhost",
						Port: "8000",
					},
					Logger: LoggerConfig{
						Level: INFO,
						Mode:  "",
					},
				}
			},
			expectError:  true,
			errorMessage: "missing logger mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			config := tc.setupConfig()

			// Test
			err := validateConfig(config)

			// Verify
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogLevel(t *testing.T) {
	t.Run("Valid Log Levels", func(t *testing.T) {
		// Setup
		resetState()
		removeTestConfigFile(t)

		testCases := []struct {
			level         string
			expectedLevel LogLevel
		}{
			{"info", INFO},
			{"warn", WARN},
			{"error", ERROR},
			{"critical", CRITICAL},
		}

		for _, tc := range testCases {
			t.Run(tc.level, func(t *testing.T) {
				// Set environment variable for logger level
				os.Setenv("RCMETERING_LOGGER_LEVEL", tc.level)
				defer os.Unsetenv("RCMETERING_LOGGER_LEVEL")

				// Test
				config, err := LoadConfig()

				// Verify
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLevel, config.Logger.Level)
			})
		}
	})
}

func TestInitializeViper(t *testing.T) {
	t.Run("With Valid Config File", func(t *testing.T) {
		// Setup
		resetState()
		configContent := `
SERVER_HOST=localhost
SERVER_PORT=8000
`
		createTestConfigFile(t, configContent)
		defer removeTestConfigFile(t)

		// Test
		err := initializeViper()

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, "localhost", viper.GetString("SERVER_HOST"))
		assert.Equal(t, "8000", viper.GetString("SERVER_PORT"))
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Invalid Log Level", func(t *testing.T) {
		// Setup
		resetState()
		os.Setenv("RCMETERING_LOGGER_LEVEL", "invalid_level")
		defer os.Unsetenv("RCMETERING_LOGGER_LEVEL")

		// Test
		config, err := LoadConfig()

		// Verify
		require.NoError(t, err)                                         // Should not error on invalid log level
		assert.Equal(t, LogLevel("invalid_level"), config.Logger.Level) // Should use the provided value even if not in enums
	})

	t.Run("Malformed Config File", func(t *testing.T) {
		// Setup
		resetState()
		configContent := `
SERVER_HOST:localhost
This is not a valid config file
`
		createTestConfigFile(t, configContent)
		defer removeTestConfigFile(t)

		// Test
		config, err := LoadConfig()

		// Verify - should fall back to defaults
		require.NoError(t, err)
		assert.Equal(t, "localhost", config.Server.Host)
		assert.Equal(t, "8000", config.Server.Port)
	})
}
