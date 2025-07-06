package main

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"

	"phpservermanager/internal/app"
	"phpservermanager/internal/config"
	"phpservermanager/internal/handler"
	"phpservermanager/internal/middleware"
)

//go:embed web/static
var staticFS embed.FS

func main() {
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.yaml")

	// Ensure the config directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}
	}

	// Check if config file exists, if not, run setup
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("No configuration file found. Starting first-time setup...")
		if err := setupConfig(configPath); err != nil {
			log.Fatalf("Failed to complete setup: %v", err)
		}
		fmt.Println("Setup complete. Starting PHP Server Manager...")
	}

	// Load configuration
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the App
	application := app.NewApp(cfg)
	application.Startup(context.Background())
	defer application.Shutdown(context.Background())

	// Initialize the handlers
	h := handler.NewHandler(application)

	// Create router
	r := mux.NewRouter()

	// Create a new auth middleware
	authMiddleware := middleware.Auth(cfg.Auth)

	// API endpoints
	api := r.PathPrefix("/api").Subrouter()
	api.Use(CORSMiddleware)
	api.Use(authMiddleware)
	api.HandleFunc("/servers", h.HandleGetServers).Methods("GET")
	api.HandleFunc("/servers", h.HandleCreateServer).Methods("POST")
	api.HandleFunc("/servers/{id}", h.HandleUpdateServer).Methods("PUT")
	api.HandleFunc("/servers/{id}", h.HandleDeleteServer).Methods("DELETE")
	api.HandleFunc("/servers/{id}/start", h.HandleStartServer).Methods("POST")
	api.HandleFunc("/servers/{id}/stop", h.HandleStopServer).Methods("POST")
	api.HandleFunc("/servers/{id}/status", h.HandleServerStatus).Methods("GET")
	api.HandleFunc("/settings", h.HandleGetServerSettings).Methods("GET")
	api.HandleFunc("/settings", h.HandleUpdateServerSettings).Methods("PUT")
	api.HandleFunc("/auth", h.HandleUpdateAuth).Methods("PUT")
	// api.HandleFunc("/acme/status", h.HandleGetACMEStatus).Methods("GET")
	// api.HandleFunc("/acme/settings", h.HandleUpdateACMESettings).Methods("PUT")
	// api.HandleFunc("/acme/renew", h.HandleRenewACME).Methods("POST")

	// Static files
	staticContent, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		log.Fatal(err)
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticContent)))

	// Start web server
	bindAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("PHP Server Manager is running at http://%s\n", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, r))
}

func getConfigDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "phpservermanager")
	case "linux":
		return filepath.Join("/etc", "phpservermanager")
	default:
		return filepath.Join(".", "phpservermanager") // Fallback for unknown OS
	}
}

func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setupConfig(path string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter initial username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter initial password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	fmt.Printf("Enter server host (default: 0.0.0.0): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "0.0.0.0"
	}

	fmt.Printf("Enter server port (default: 8080): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Enter path for servers data (default: %s/servers.json): ", getConfigDir())
	serversConfigPath, _ := reader.ReadString('\n')
	serversConfigPath = strings.TrimSpace(serversConfigPath)
	if serversConfigPath == "" {
		serversConfigPath = filepath.Join(getConfigDir(), "servers.json")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	cfg := config.Config{
		Server: config.ServerConfig{
			Host: host,
			Port: port,
		},
		Auth: config.Auth{
			Username: username,
			PasswordHash: string(hashedPassword),
		},
		ServersConfigPath: serversConfigPath,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CORSMiddleware adds CORS headers to the response
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}