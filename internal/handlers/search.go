package handlers

import (
	"encoding/json"
	"jobseek-web-be/internal/search"
	"log"
	"net/http"
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

	// Map request to service params
	params := search.SearchParams{
		Keyword:       req.Keyword,
		Country:       req.Country,
		Location:      req.Location,
		LocalLanguage: req.LocalLanguage,
		ResultsWanted: req.ResultsWanted,
	}

	results, err := search.ExecuteSearch(params)
	if err != nil {
		log.Printf("Search failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
