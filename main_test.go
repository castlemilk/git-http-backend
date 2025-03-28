package main

import (
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getRandomPort returns a random available port
func getRandomPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// setupTestRepo creates and initializes a test Git repository
func setupTestRepo(t *testing.T, tmpDir string) {
	t.Helper()

	// Set Git environment variables
	os.Setenv("GIT_HTTP_BACKEND_ENABLE_RECEIVE_PACK", "true")
	os.Setenv("GIT_HTTP_BACKEND_ENABLE_UPLOAD_PACK", "true")
	os.Setenv("GIT_HTTP_EXPORT_ALL", "true")

	// Create a test git repository
	repoPath := filepath.Join(tmpDir, "test-repo.git")
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Initialize bare git repository
	cmd := exec.Command("git", "init", "--bare", repoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to init bare repo: %s", string(output))
		t.Fatal(err)
	}

	// Create a temporary working directory
	workDir := filepath.Join(tmpDir, "work")
	err = os.MkdirAll(workDir, 0755)
	require.NoError(t, err)

	// Initialize working repository
	cmd = exec.Command("git", "init", workDir)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to init working repo: %s", string(output))
		t.Fatal(err)
	}

	// Configure git in working directory
	cmd = exec.Command("git", "-C", workDir, "config", "user.name", "Test User")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to configure user.name: %s", string(output))
		t.Fatal(err)
	}

	cmd = exec.Command("git", "-C", workDir, "config", "user.email", "test@example.com")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to configure user.email: %s", string(output))
		t.Fatal(err)
	}

	// Create a test file
	testFile := filepath.Join(workDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Add and commit the file
	cmd = exec.Command("git", "-C", workDir, "add", "test.txt")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to add file: %s", string(output))
		t.Fatal(err)
	}

	cmd = exec.Command("git", "-C", workDir, "commit", "-m", "Initial commit")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to commit: %s", string(output))
		t.Fatal(err)
	}

	// Get the current branch name
	cmd = exec.Command("git", "-C", workDir, "branch", "--show-current")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to get branch name: %s", string(output))
		t.Fatal(err)
	}
	branchName := strings.TrimSpace(string(output))
	if branchName == "" {
		branchName = "main" // Default to main if no branch name is found
	}

	// Push to the bare repository
	cmd = exec.Command("git", "-C", workDir, "push", repoPath, branchName+":"+branchName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to push: %s", string(output))
		t.Fatal(err)
	}
}

func TestValidateCredentials(t *testing.T) {
	// Reset viper configuration for tests
	viper.Reset()

	// Set test configuration
	viper.Set("username", "testuser")
	viper.Set("password", "testpass")

	// Test valid credentials
	assert.True(t, validateCredentials("testuser", "testpass"))

	// Test invalid credentials
	assert.False(t, validateCredentials("wronguser", "testpass"))
	assert.False(t, validateCredentials("testuser", "wrongpass"))
	assert.False(t, validateCredentials("wronguser", "wrongpass"))
}

func TestGitHTTPBackendServer(t *testing.T) {
	// Reset viper configuration for tests
	viper.Reset()

	// Set test configuration
	viper.Set("username", "testuser")
	viper.Set("password", "testpass")
	viper.Set("port", 3000)
	viper.Set("server-temp-dir", os.TempDir()+"/git-test")

	// Create temporary directory for test
	tmpDir := t.TempDir()
	setupTestRepo(t, tmpDir)

	// Start server
	server, err := startGitHTTPBackendServer(tmpDir, 3000)
	require.NoError(t, err)
	defer server.Close()

	// Test basic auth
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:3000/test-repo.git/info/refs", nil)
	require.NoError(t, err)

	// Test without auth
	resp, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test with auth
	req.SetBasicAuth("testuser", "testpass")
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServerEnvironmentVariables(t *testing.T) {
	// Reset viper configuration for tests
	viper.Reset()

	// Set test environment variables
	os.Setenv("GIT_USERNAME", "envuser")
	os.Setenv("GIT_PASSWORD", "envpass")
	os.Setenv("GIT_PORT", "8080")
	os.Setenv("GIT_SERVER_TEMP_DIR", "/tmp/git-env-test")

	// Reinitialize configuration
	InitConfig()

	// Verify environment variables are set correctly
	assert.Equal(t, "envuser", config.Username)
	assert.Equal(t, "envpass", config.Password)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "/tmp/git-env-test", config.ServerTempDir)
}
