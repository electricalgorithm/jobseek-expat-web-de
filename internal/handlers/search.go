package handlers

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

type SearchRequest struct {
	Keyword       string `json:"keyword"`
	Country       string `json:"country"`
	Location      string `json:"location"`
	LocalLanguage string `json:"local_language"`
	ResultsWanted int    `json:"results_wanted"`
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// NOTE: In a real implementation with plans, we would check the user's plan here.
	// For now, Basic Plan allows instant search, which is what this is.
	// We might check if the user is authenticated via middleware, but for the 'free partial access' or 'basic'
	// we keep it open or require the token.

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

	// Capture stderr separately
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error running command: %v, Stderr: %s", err, stderr.String())
		http.Error(w, fmt.Sprintf("Error executing search: %s", stderr.String()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

func getJobSeekPath() string {
	path, err := exec.LookPath("jobseek-expat")
	if err == nil {
		return path
	}
	home, _ := os.UserHomeDir()
	fallback := filepath.Join(home, "Library/Python/3.14/bin/jobseek-expat")
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}
	// Fallback for older python versions or different setups
	fallback312 := filepath.Join(home, "Library/Python/3.12/bin/jobseek-expat")
	if _, err := os.Stat(fallback312); err == nil {
		return fallback312
	}

	return "jobseek-expat"
}
