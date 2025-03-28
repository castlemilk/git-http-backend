package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Port          int    `mapstructure:"port"`
	ServerTempDir string `mapstructure:"server-temp-dir"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
}

var config Config
var flagSet *pflag.FlagSet

// InitConfig initializes the configuration using Viper
func InitConfig() {
	// Reset viper and create new flag set
	viper.Reset()
	flagSet = pflag.NewFlagSet("git-http-backend", pflag.ExitOnError)

	// Set default values
	viper.SetDefault("port", 3000)
	viper.SetDefault("server-temp-dir", os.TempDir()+"/git")
	viper.SetDefault("username", "testuser")
	viper.SetDefault("password", "testpass")

	// Enable environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GIT")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Define flags
	flagSet.Int("port", 3000, "port to listen on")
	flagSet.String("server-temp-dir", os.TempDir()+"/git", "temp dir to use for the server")
	flagSet.String("username", "testuser", "username to use for the server")
	flagSet.String("password", "testpass", "password to use for the server")

	// Parse flags if not in test mode
	if !strings.HasSuffix(os.Args[0], ".test") {
		flagSet.Parse(os.Args[1:])
	}

	// Bind flags to viper
	viper.BindPFlags(flagSet)

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %s", err)
		}
	}

	// Unmarshal configuration
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}
}

func init() {
	InitConfig()
}

// basicAuthMiddleware creates a middleware that handles HTTP Basic Authentication
func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			unauthorized(w)
			return
		}

		if !validateCredentials(username, password) {
			unauthorized(w)
			return
		}

		// Set the authenticated username in the request header for downstream handlers
		r.Header.Set("X-Remote-User", username)
		next.ServeHTTP(w, r)
	})
}

// validateCredentials checks if the provided credentials match
func validateCredentials(username, password string) bool {
	return username == config.Username && password == config.Password
}

// unauthorized sends an unauthorized response
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func startGitHTTPBackendServer(serverTempDir string, port int) (*http.Server, error) {
	log.Printf("Starting Git HTTP backend server with temp dir: %s", serverTempDir)

	// Find git-http-backend path
	gitExecPath := os.Getenv("GIT_EXEC_PATH")
	if gitExecPath == "" {
		gitExecPathCmd := exec.Command("git", "--exec-path")
		gitExecPathOutput, err := gitExecPathCmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get git exec path: %v", err)
		}
		gitExecPath = strings.TrimSpace(string(gitExecPathOutput))
	}
	gitHTTPBackend := filepath.Join(gitExecPath, "git-http-backend")
	log.Printf("Using git-http-backend at: %s", gitHTTPBackend)

	// Verify that git-http-backend exists
	if _, err := os.Stat(gitHTTPBackend); os.IsNotExist(err) {
		// Try alternative paths
		alternativePaths := []string{
			"/usr/libexec/git-core/git-http-backend",
			"/usr/lib/git-core/git-http-backend",
			"/usr/local/libexec/git-core/git-http-backend",
		}
		for _, path := range alternativePaths {
			if _, err := os.Stat(path); err == nil {
				gitHTTPBackend = path
				log.Printf("Found git-http-backend at alternative path: %s", path)
				break
			}
		}
		if _, err := os.Stat(gitHTTPBackend); os.IsNotExist(err) {
			return nil, fmt.Errorf("git-http-backend not found in any of the expected locations")
		}
	}

	// Create a new ServeMux for this server instance
	mux := http.NewServeMux()

	// Create git handler
	gitHandler := &cgi.Handler{
		Path: gitHTTPBackend,
		Dir:  serverTempDir,
		Env: []string{
			"GIT_PROJECT_ROOT=" + serverTempDir,
			"GIT_HTTP_EXPORT_ALL=true",
			"GIT_HTTP_MAX_REQUEST_BUFFER=1000M",
		},
	}

	// Add the git handler to the mux
	mux.Handle("/", basicAuthMiddleware(gitHandler))

	// Create the server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return server, nil
}

func main() {
	server, err := startGitHTTPBackendServer(config.ServerTempDir, config.Port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("Server started on port %d", config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
