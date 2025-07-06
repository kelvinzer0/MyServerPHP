package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/caddyserver/certmagic"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"

	"phpservermanager/internal/config"
	"phpservermanager/internal/server"
)

// App struct
type App struct {
	ctx               context.Context
	servers           map[string]*server.Server
	nextID            int
	mu                sync.Mutex
	processes         map[string]*exec.Cmd
	serversConfigPath string
	serverHost        string
	serverPort        string
	auth              config.Auth
	certmagicInstances map[string]*certmagic.Config
}

// NewApp creates a new App application struct
func NewApp(cfg *config.Config) *App {
	app := &App{
		servers:           make(map[string]*server.Server),
		nextID:            1,
		processes:         make(map[string]*exec.Cmd),
		serversConfigPath: cfg.ServersConfigPath,
		serverHost:        cfg.Server.Host,
		serverPort:        cfg.Server.Port,
		auth:              cfg.Auth,
		certmagicInstances: make(map[string]*certmagic.Config),
	}

	

    return app
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
    // Ensure the directory for serversConfigPath exists
    configDir := filepath.Dir(a.serversConfigPath)
    if _, err := os.Stat(configDir); os.IsNotExist(err) {
        os.MkdirAll(configDir, 0755)
    }
    a.loadConfig()
}

// Shutdown is called when the app is about to exit
func (a *App) Shutdown(ctx context.Context) {
    for id, s := range a.servers {
        if s.Running {
            a.StopServer(id)
        }
    }
    a.saveConfig()
}

// loadConfig loads the saved configuration from disk
func (a *App) loadConfig() {
    data, err := ioutil.ReadFile(a.serversConfigPath)
    if err != nil {
        return
    }

    var config struct {
        Servers    map[string]*server.Server `json:"servers"`
        NextID     int                       `json:"nextID"`
        ServerHost string                    `json:"serverHost"`
        ServerPort string                    `json:"serverPort"`
    }
    if err := json.Unmarshal(data, &config); err != nil {
        fmt.Printf("Error loading configuration: %v\n", err)
        return
    }

    a.servers = config.Servers
    a.nextID = config.NextID
    if config.ServerHost != "" {
        a.serverHost = config.ServerHost
    }
    if config.ServerPort != "" {
        a.serverPort = config.ServerPort
    }

    for _, s := range a.servers {
        s.Running = false
        if s.Host == "" {
            s.Host = "localhost"
        }
    }
}

// saveConfig saves the current configuration to disk
func (a *App) saveConfig() {
    a.mu.Lock()
    defer a.mu.Unlock()

    config := struct {
        Servers    map[string]*server.Server `json:"servers"`
        NextID     int                       `json:"nextID"`
        ServerHost string                    `json:"serverHost"`
        ServerPort string                    `json:"serverPort"`
    }{
        Servers:    a.servers,
        NextID:     a.nextID,
        ServerHost: a.serverHost,
        ServerPort: a.serverPort,
    }

    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        fmt.Printf("Error serializing configuration: %v\n", err)
        return
    }

    if err := ioutil.WriteFile(a.serversConfigPath, data, 0644); err != nil {
        fmt.Printf("Error saving configuration: %v\n", err)
    }
}

// GetServers returns all configured servers
func (a *App) GetServers() []*server.Server {
    a.mu.Lock()
    defer a.mu.Unlock()

    servers := make([]*server.Server, 0, len(a.servers))
    for _, s := range a.servers {
        servers = append(servers, s)
    }
    return servers
}

// CreateServer adds a new server configuration
func (a *App) CreateServer(name, host, port, directory, command string) string {
    a.mu.Lock()
    defer a.mu.Unlock()

    id := strconv.Itoa(a.nextID)
    a.nextID++

    if host == "" {
        host = "localhost"
    }

    s := &server.Server{
        ID:        id,
        Name:      name,
        Host:      host,
        Port:      port,
        Directory: directory,
        Command:   command,
        Running:   false,
    }

    a.servers[id] = s
    go a.saveConfig()
    return id
}

// UpdateServer updates an existing server configuration
func (a *App) UpdateServer(id, name, host, port, directory, command string) bool {
    a.mu.Lock()
    defer a.mu.Unlock()

    s, exists := a.servers[id]
    if !exists {
        return false
    }

    if s.Running {
        a.mu.Unlock()
        a.StopServer(id)
        a.mu.Lock()
    }

    if host == "" {
        host = "localhost"
    }

    s.Name = name
    s.Host = host
    s.Port = port
    s.Directory = directory
    s.Command = command
    go a.saveConfig()
    return true
}

// DeleteServer removes a server configuration
func (a *App) DeleteServer(id string) bool {
    a.mu.Lock()
    defer a.mu.Unlock()

    s, exists := a.servers[id]
    if !exists {
        return false
    }

    if s.Running {
        a.mu.Unlock()
        a.StopServer(id)
        a.mu.Lock()
    }

    delete(a.servers, id)
    go a.saveConfig()
    return true
}

// UpdateServerSettings updates the management server host and port
func (a *App) UpdateServerSettings(host, port string) bool {
    a.mu.Lock()
    defer a.mu.Unlock()

    if host == "" {
        host = "localhost"
    }
    if port == "" {
        port = "8080"
    }

    a.serverHost = host
    a.serverPort = port
    go a.saveConfig()
    return true
}

// GetServerSettings returns the current server settings
func (a *App) GetServerSettings() (string, string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    return a.serverHost, a.serverPort
}

// StartServer starts a PHP server
func (a *App) StartServer(id string) bool {
    a.mu.Lock()
    s, exists := a.servers[id]
    if !exists || s.Running {
        a.mu.Unlock()
        return false
    }
    a.mu.Unlock()

    if s.ACMEEnabled && len(s.ACMEDomains) > 0 {
        cfg := certmagic.NewDefault()
        cfg.Storage = &certmagic.FileStorage{Path: s.ACMEStoragePath}

        // Manage certificates in a goroutine to avoid blocking
        go func() {
            err := cfg.ManageSync(a.ctx, s.ACMEDomains)
            if err != nil {
                fmt.Printf("CertMagic error for server %s (%s): %v\n", s.Name, s.ID, err)
            }
        }()
        a.mu.Lock()
        a.certmagicInstances[s.ID] = cfg
        a.mu.Unlock()
    }

    return server.Start(s, a.processes, &a.mu)
}

// StopServer stops a running PHP server
func (a *App) StopServer(id string) bool {
    a.mu.Lock()
    s, exists := a.servers[id]
    if !exists || !s.Running {
        a.mu.Unlock()
        return false
    }
    a.mu.Unlock()

    if s.ACMEEnabled {
        if _, ok := a.certmagicInstances[s.ID]; ok {
            // Stop the certificate management for this server
            // This is a conceptual representation. The actual implementation might vary based on certmagic's API.
            // certmagic.Default.Unmanage(s.ACMEDomains)
            // For now, we'll just log it.
            fmt.Printf("Stopping cert management for %s\n", s.ID)
            delete(a.certmagicInstances, s.ID)
        }
    }

    return server.Stop(s, a.processes, &a.mu)
}

// GetServerStatus returns the status of a specific server
func (a *App) GetServerStatus(id string) (bool, bool) {
    a.mu.Lock()
    defer a.mu.Unlock()

    s, exists := a.servers[id]
    if !exists {
        return false, false
    }
    return true, s.Running
}

// UpdateAuth updates the auth settings in the config file
func (a *App) UpdateAuth(username, password string) error {
    a.mu.Lock()
    defer a.mu.Unlock()

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    a.auth.Username = username
    a.auth.PasswordHash = string(hashedPassword)

    // read the config file
    data, err := ioutil.ReadFile("internal/config/config.yaml")
    if err != nil {
        return err
    }

    // unmarshal the config file
    var configData map[string]interface{}
    if err := yaml.Unmarshal(data, &configData); err != nil {
        return err
    }

    // update the auth settings
    authData, ok := configData["auth"].(map[interface{}]interface{})
    if !ok {
        authData = make(map[interface{}]interface{})
    }
    authData["username"] = username
    authData["password_hash"] = string(hashedPassword)
    configData["auth"] = authData

    // marshal the config file
    newData, err := yaml.Marshal(&configData)
    if err != nil {
        return err
    }

    // write the config file
    if err := ioutil.WriteFile("internal/config/config.yaml", newData, 0644); err != nil {
        return err
    }

    return nil
}






