package server

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

// Server represents a PHP server configuration
type Server struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Host            string `json:"host"`
	Port            string `json:"port"`
	Directory       string `json:"directory"`
	Command         string `json:"command"`
	Running         bool   `json:"running"`
	ACMEEnabled     bool   `json:"acme_enabled"`
	ACMECertEmail   string `json:"acme_cert_email"`
	ACMEDomains     []string `json:"acme_domains"`
	ACMEStoragePath string `json:"acme_storage_path"`
}

// Start starts a PHP server
func Start(s *Server, processes map[string]*exec.Cmd, mu *sync.Mutex) bool {
	var command string
	bindHost := formatHostForBinding(s.Host)
	listenAddr := bindHost + ":" + s.Port
	if s.ACMEEnabled {
		listenAddr = "https://" + listenAddr
	}
	if s.Command != "" {
		command = s.Command
		command = strings.ReplaceAll(command, "{host}", s.Host)
		command = strings.ReplaceAll(command, "{port}", s.Port)
		command = strings.ReplaceAll(command, "{directory}", s.Directory)
		command = strings.ReplaceAll(command, "{bind_host}", bindHost)
		command = strings.ReplaceAll(command, "{listen_addr}", listenAddr)
	} else {
		command = fmt.Sprintf("frankenphp php-server --listen %s -r %s", listenAddr, s.Directory)
	}

	os.Setenv("PATH", "/usr/local/bin:"+os.Getenv("PATH")) // Tetap untuk Linux/macOS

	username := getCurrentUsername()
	fullCommand := fmt.Sprintf("sudo -u %s /bin/bash -c '%s'", username, command)
	cmd := exec.Command("/bin/bash", "-c", fullCommand)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Dir, _ = os.Getwd()

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return false
	}

	mu.Lock()
	processes[s.ID] = cmd
	s.Running = true
	mu.Unlock()

	go func() {
		cmd.Wait()
		mu.Lock()
		delete(processes, s.ID)
		s.Running = false
		mu.Unlock()
	}()

	return true
}

// Stop stops a running PHP server
func Stop(s *Server, processes map[string]*exec.Cmd, mu *sync.Mutex) bool {
	mu.Lock()
	cmd, exists := processes[s.ID]
	if !exists {
		s.Running = false
		mu.Unlock()
		return true
	}
	mu.Unlock()

	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
		fmt.Printf("Error stopping server: %v\n", err)
		return false
	}

	mu.Lock()
	delete(processes, s.ID)
	s.Running = false
	mu.Unlock()

	return true
}

func getCurrentUsername() string {
	user, err := os.UserHomeDir()
	if err != nil {
		return "root" // Fallback for non-Windows
	}
	return filepath.Base(user)
}

func formatHostForBinding(host string) string {
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		if net.ParseIP(host) != nil {
			return "[" + host + "]"
		}
	}
	return host
}