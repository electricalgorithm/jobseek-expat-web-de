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
}

func ExecuteSearch(params SearchParams) ([]interface{}, error) {
	// Default values
	if params.Country == "" {
		params.Country = "Germany"
	}
	resultsWanted := "10"
	if params.ResultsWanted > 0 {
		resultsWanted = fmt.Sprintf("%d", params.ResultsWanted)
	}

	// Construct arguments - we assume "search" command is implicit or entry point?
	// The original handler called `exec.Command(cmdPath, args...)` where args started with Keyword (argument 1).
	// Let's verify how the CLI works.
	// The `jobseek-expat` command takes arguments directly?
	// Looking at handlers/search.go: args := []string{req.Keyword, ...}
	// So `jobseek-expat <keyword> ...`
	// Wait, viewed_file 1 said "Start of search command definition".
	// The CLI entry point `jobseek-expat` might need a `search` subcommand if it's a Typer app with multiple commands.
	// But the handler uses `args := []string{req.Keyword, ...}`. This suggests `jobseek-expat keyword --options`.
	// I will stick to what the handler was doing.

	args := []string{params.Keyword, "--country", params.Country, "--output", "json", "--results-wanted", resultsWanted}

	if params.Location != "" {
		args = append(args, "--location", params.Location)
	}
	if params.LocalLanguage != "" {
		args = append(args, "--local-language", params.LocalLanguage)
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
