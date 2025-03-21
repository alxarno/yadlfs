//nolint:paralleltest
package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileConfigLoad(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Create a YAML file with test configuration
	yamlContent := `
yandexDiskOauthToken: "test_oauth_token"
yandexDiskProjectFolder: "/test/project/folder"
`
	yamlFilePath := filepath.Join(tempDir, ".yadlfs.yaml")
	err := os.WriteFile(yamlFilePath, []byte(yamlContent), 0600)
	require.NoError(t, err, "Failed to create YAML file")

	// Change the working directory to the temporary directory
	t.Chdir(tempDir)

	// Load the configuration
	config, err := LoadConfig()
	require.NoError(t, err, "Failed to load config from YAML file")

	// Verify the configuration values
	require.Equal(t, "test_oauth_token", config.YandexDiskOAuthToken, "YandexDiskOAuthToken mismatch")
	require.Equal(t, "/test/project/folder", config.YandexDiskProjectFolder, "YandexDiskProjectFolder mismatch")
}

func TestEnvConfigLoad(t *testing.T) {
	// Set up environment variables for the test
	t.Setenv("YANDEX_DISK_OAUTH_TOKEN", "test_oauth_token")
	t.Setenv("YANDEX_DISK_PROJECT_FOLDER", "/test/project/folder")

	// Load the configuration
	config, err := LoadConfig()
	require.NoError(t, err, "Failed to load config from environment variables")

	// Verify the configuration values
	require.Equal(t, "test_oauth_token", config.YandexDiskOAuthToken, "YandexDiskOAuthToken mismatch")
	require.Equal(t, "/test/project/folder", config.YandexDiskProjectFolder, "YandexDiskProjectFolder mismatch")
}

func TestConfigFailedLoad(t *testing.T) {
	// Test case: Invalid YAML file
	t.Run("InvalidYAMLFile", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()

		// Create an invalid YAML file
		invalidYAMLContent := `
invalid_yaml_content
`
		yamlFilePath := filepath.Join(tempDir, ".yadlfs.yaml")
		err := os.WriteFile(yamlFilePath, []byte(invalidYAMLContent), 0600)
		require.NoError(t, err, "Failed to create invalid YAML file")

		t.Chdir(tempDir)

		// Attempt to load the configuration
		_, err = LoadConfig()
		require.Error(t, err, "Expected an error when YAML file is invalid")
	})
}
