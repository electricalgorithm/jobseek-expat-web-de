package handlers

import (
	"encoding/json"
	"jobseek-web-be/internal/auth"
	"jobseek-web-be/internal/search"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

	// Require authentication for search
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return auth.SecretKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

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
