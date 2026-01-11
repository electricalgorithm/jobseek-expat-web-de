package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// API Routes
	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/search", searchHandler)

	// Serve Static Files (Frontend)
	// assuming dist is in ../jobseek-web-fe/dist based on previous setup
	fs := http.FileServer(http.Dir("../jobseek-web-fe/dist"))
	http.Handle("/", fs)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

type SearchRequest struct {
	Keyword       string `json:"keyword"`
	Country       string `json:"country"`
	Location      string `json:"location"`
	LocalLanguage string `json:"local_language"`
	ResultsWanted int    `json:"results_wanted"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "jobseek-web-be"})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Default values
	if req.Country == "" {
		req.Country = "Germany"
	}
	resultsWanted := "10"
	if req.ResultsWanted > 0 {
		resultsWanted = fmt.Sprintf("%d", req.ResultsWanted)
	}

	// Construct arguments
	args := []string{req.Keyword, "--country", req.Country, "--output", "json", "--results-wanted", resultsWanted}

	if req.Location != "" {
		args = append(args, "--location", req.Location)
	}
	if req.LocalLanguage != "" {
		args = append(args, "--local-language", req.LocalLanguage)
	}

	log.Printf("Running search: jobseek-expat %v", args)

	// Execute CLI
	cmdPath := getJobSeekPath()
	cmd := exec.Command(cmdPath, args...)

	// Capture stderr separately for logging, so it doesn't pollute the JSON response
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error running command: %v, Stderr: %s", err, stderr.String())
		// Return a JSON error object instead of plain text if possible, or 500
		http.Error(w, fmt.Sprintf("Error executing search: %s", stderr.String()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

func getJobSeekPath() string {
	// 1. Check if 'jobseek-expat' is in PATH
	path, err := exec.LookPath("jobseek-expat")
	if err == nil {
		return path
	}

	// 2. Fallback to known user bin path on macOS (common for pip install --user)
	// Adjust version if needed, or check multiple.
	home, _ := os.UserHomeDir()
	fallback := filepath.Join(home, "Library/Python/3.14/bin/jobseek-expat")
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}

	return "jobseek-expat" // Default and hope for the best
}
