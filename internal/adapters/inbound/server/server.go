package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"takeout_services/internal/adapters/outbound/repository"
)

// ToggleRequest represents the POST payload to check/uncheck a service.
type ToggleRequest struct {
	Domain  string `json:"domain"`
	Deleted bool   `json:"deleted"`
}

// StartServer launches the Go-native HTTP server to host the dashboard.
func StartServer(port int, reportPath, statePath string) error {
	stateRepo := repository.NewFileStateRepository(statePath)

	// API: Get list of deleted domains
	http.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Re-read file to get fresh updates
		stateRepo = repository.NewFileStateRepository(statePath)

		// Get all deleted domains
		var list []string
		// We use a temporary map iteration to find what's true
		for _, service := range stateRepo.IsDeletedList() {
			list = append(list, service)
		}

		json.NewEncoder(w).Encode(map[string][]string{
			"deleted": list,
		})
	})

	// API: Toggle deleted status
	http.HandleFunc("/api/toggle-deleted", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ToggleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err := stateRepo.Save(req.Domain, req.Deleted); err != nil {
			http.Error(w, fmt.Sprintf("Error saving state: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Static files & Dashboard Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/report.html" {
			if _, err := os.Stat(reportPath); os.IsNotExist(err) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, `
					<div style="font-family: sans-serif; max-width: 600px; margin: 4rem auto; padding: 2rem; border: 1px solid #e2e8f0; border-radius: 1rem; text-align: center;">
						<h2>No report found.</h2>
						<p>Please run the digital footprint analysis first to generate the report dashboard:</p>
						<code style="background: #f1f5f9; padding: 0.5rem 1rem; border-radius: 0.5rem; display: block; margin: 1rem 0;">
							./takeout-parser --input "Takeout/Почта/Вся почта, включая _Спам_ и _Корзину_.mbox" --detect
						</code>
					</div>
				`)
				return
			}
			http.ServeFile(w, r, reportPath)
			return
		}

		// Fallback to serving general files in the same directory (e.g. footprint.json)
		dir := filepath.Dir(reportPath)
		http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
	})

	fmt.Printf("Starting local Go Web Server at: http://localhost:%d/\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
