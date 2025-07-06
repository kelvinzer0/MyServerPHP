package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"phpservermanager/internal/app"
)

// Handler struct
type Handler struct {
	App *app.App
}

// NewHandler creates a new Handler
func NewHandler(a *app.App) *Handler {
	return &Handler{App: a}
}

// HandleGetServers handles the GET /api/servers endpoint
func (h *Handler) HandleGetServers(w http.ResponseWriter, r *http.Request) {
	servers := h.App.GetServers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// HandleCreateServer handles the POST /api/servers endpoint
func (h *Handler) HandleCreateServer(w http.ResponseWriter, r *http.Request) {
	var serverData struct {
		Name      string `json:"name"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		Directory string `json:"directory"`
		Command   string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&serverData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if serverData.Name == "" || serverData.Port == "" || serverData.Directory == "" {
		http.Error(w, "Name, port, and directory are required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(serverData.Port); err != nil {
		http.Error(w, "Port must be a number", http.StatusBadRequest)
		return
	}

	if serverData.Host != "" && !validateHost(serverData.Host) {
		http.Error(w, "Invalid host format", http.StatusBadRequest)
		return
	}

	id := h.App.CreateServer(serverData.Name, serverData.Host, serverData.Port, serverData.Directory, serverData.Command)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// HandleUpdateServer handles the PUT /api/servers/{id} endpoint
func (h *Handler) HandleUpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var serverData struct {
		Name      string `json:"name"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		Directory string `json:"directory"`
		Command   string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&serverData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if serverData.Name == "" || serverData.Port == "" || serverData.Directory == "" {
		http.Error(w, "Name, port, and directory are required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(serverData.Port); err != nil {
		http.Error(w, "Port must be a number", http.StatusBadRequest)
		return
	}

	if serverData.Host != "" && !validateHost(serverData.Host) {
		http.Error(w, "Invalid host format", http.StatusBadRequest)
		return
	}

	success := h.App.UpdateServer(id, serverData.Name, serverData.Host, serverData.Port, serverData.Directory, serverData.Command)
	if !success {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleDeleteServer handles the DELETE /api/servers/{id} endpoint
func (h *Handler) HandleDeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	success := h.App.DeleteServer(id)
	if !success {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleStartServer handles the POST /api/servers/{id}/start endpoint
func (h *Handler) HandleStartServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	success := h.App.StartServer(id)
	if !success {
		http.Error(w, "Failed to start server or server is already running", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleStopServer handles the POST /api/servers/{id}/stop endpoint
func (h *Handler) HandleStopServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	success := h.App.StopServer(id)
	if !success {
		http.Error(w, "Failed to stop server or server is already stopped", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleServerStatus handles the GET /api/servers/{id}/status endpoint
func (h *Handler) HandleServerStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	exists, running := h.App.GetServerStatus(id)
	if !exists {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"running": running})
}

// HandleGetServerSettings handles the GET /api/settings endpoint
func (h *Handler) HandleGetServerSettings(w http.ResponseWriter, r *http.Request) {
	host, port := h.App.GetServerSettings()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"host": host,
		"port": port,
	})
}

// HandleUpdateServerSettings handles the PUT /api/settings endpoint
func (h *Handler) HandleUpdateServerSettings(w http.ResponseWriter, r *http.Request) {
	var settingsData struct {
		Host string `json:"host"`
		Port string `json:"port"`
	}

	if err := json.NewDecoder(r.Body).Decode(&settingsData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if settingsData.Port != "" {
		if _, err := strconv.Atoi(settingsData.Port); err != nil {
			http.Error(w, "Port must be a number", http.StatusBadRequest)
			return
		}
	}

	if settingsData.Host != "" && !validateHost(settingsData.Host) {
		http.Error(w, "Invalid host format", http.StatusBadRequest)
		return
	}

	success := h.App.UpdateServerSettings(settingsData.Host, settingsData.Port)
	if !success {
		http.Error(w, "Failed to update server settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settings updated successfully. Restart the application to apply changes."})
}

// HandleUpdateAuth handles the PUT /api/auth endpoint
func (h *Handler) HandleUpdateAuth(w http.ResponseWriter, r *http.Request) {
	var authData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if authData.Username == "" || authData.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if err := h.App.UpdateAuth(authData.Username, authData.Password); err != nil {
		http.Error(w, "Failed to update auth settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Auth settings updated successfully."})
}

// HandleGetACMEStatus handles the GET /api/acme/status endpoint
// func (h *Handler) HandleGetACMEStatus(w http.ResponseWriter, r *http.Request) {
// 	status, err := h.App.GetACMEStatus()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]bool{"enabled": status})
// }

// HandleUpdateACMESettings handles the PUT /api/acme/settings endpoint
// func (h *Handler) HandleUpdateACMESettings(w http.ResponseWriter, r *http.Request) {
// 	var acmeData struct {
// 		Enabled bool     `json:"enabled"`
// 		Email   string   `json:"email"`
// 		Domains []string `json:"domains"`
// 	}
// 
// 	if err := json.NewDecoder(r.Body).Decode(&acmeData); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 
// 	if acmeData.Enabled && (acmeData.Email == "" || len(acmeData.Domains) == 0) {
// 		http.Error(w, "Email and domains are required to enable ACME", http.StatusBadRequest)
// 		return
// 	}
// 
// 	if err := h.App.UpdateACMESettings(acmeData.Enabled, acmeData.Email, acmeData.Domains); err != nil {
// 		http.Error(w, "Failed to update ACME settings", http.StatusInternalServerError)
// 		return
// 	}
// 
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{"message": "ACME settings updated successfully. Restart the application to apply changes."})
// }

// HandleRenewACME handles the POST /api/acme/renew endpoint
// func (h *Handler) HandleRenewACME(w http.ResponseWriter, r *http.Request) {
// 	if err := h.App.RenewACME(); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{"message": "ACME certificate renewal initiated."})
// }

// ServeStatic serves static files
func ServeStatic(fs http.FileSystem) http.Handler {
	return http.FileServer(fs)
}

func validateHost(host string) bool {
	if host == "" {
		return false
	}
	if net.ParseIP(host) != nil {
		return true
	}
	if host == "localhost" || host == "0.0.0.0" || host == "::" {
		return true
	}
	if len(host) > 0 && len(host) <= 253 {
		for _, part := range strings.Split(host, ".") {
			if len(part) == 0 || len(part) > 63 {
				return false
			}
		}
		return true
	}
	return false
}
