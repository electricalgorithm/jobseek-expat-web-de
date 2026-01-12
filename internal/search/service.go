package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type SearchParams struct {
	Keyword       string
	Country       string
	Location      string
	LocalLanguage string
	ResultsWanted int
	HoursOld      int
	Exclude       string
}

func ExecuteSearch(params SearchParams) ([]interface{}, error) {
	// Default values
	if params.Country == "" {
		params.Country = "Germany"
	}
	resultsWanted := "30"
	if params.ResultsWanted > 0 {
		resultsWanted = fmt.Sprintf("%d", params.ResultsWanted)
	}

	args := []string{params.Keyword, "--country", params.Country, "--output", "json", "--results-wanted", resultsWanted}

	// Explicitly select sites (excluding Glassdoor)
	args = append(args, "--site", "linkedin", "--site", "indeed")

	if params.Location != "" {
		args = append(args, "--location", params.Location)
	}
	if params.LocalLanguage != "" {
		args = append(args, "--local-language", params.LocalLanguage)
	}
	if params.HoursOld > 0 {
		args = append(args, "--hours-old", fmt.Sprintf("%d", params.HoursOld))
	}
	if params.Exclude != "" {
		args = append(args, "--exclude", params.Exclude)
	}

	log.Printf("Running search (Service): jobseek-expat %v", args)

	// Execute CLI
	cmdPath := getJobSeekPath()
	cmd := exec.Command(cmdPath, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing search: %s (stderr: %s)", err, stderr.String())
	}

	var results []interface{}
	// The output is expected to be a JSON array
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %v", err)
	}

	return results, nil
}

func getJobSeekPath() string {
	path, err := exec.LookPath("jobseek-expat")
	if err == nil {
		return path
	}
	home, _ := os.UserHomeDir()

	// Check common user bin paths
	paths := []string{
		filepath.Join(home, "Library/Python/3.14/bin/jobseek-expat"),
		filepath.Join(home, "Library/Python/3.12/bin/jobseek-expat"),
		filepath.Join(home, ".local/bin/jobseek-expat"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "jobseek-expat"
}
