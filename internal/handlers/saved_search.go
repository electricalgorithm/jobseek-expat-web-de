package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"jobseek-web-be/internal/auth"
	"jobseek-web-be/internal/db"
	"jobseek-web-be/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

func SaveSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a DELETE request with an ID in the path
	// e.g., /api/searches/123
	if r.Method == http.MethodDelete && r.URL.Path != "/api/searches" {
		deleteSearchHandler(w, r)
		return
	}

	// Route based on method for /api/searches
	switch r.Method {
	case http.MethodGet:
		listSearchesHandler(w, r)
	case http.MethodPost:
		createSearchHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func listSearchesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Verify Auth Token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, "Invalid token email", http.StatusUnauthorized)
		return
	}

	// Get User ID
	var userID int
	err = db.DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Fetch all searches for this user
	rows, err := db.DB.Query(`
		SELECT id, keyword, country, location, language, frequency, hours_old, exclude, results_wanted, last_run 
		FROM user_searches 
		WHERE user_id = ?
		ORDER BY id DESC
	`, userID)
	if err != nil {
		http.Error(w, "Failed to fetch searches", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var searches []models.UserSearch
	for rows.Next() {
		var s models.UserSearch
		var location, language, exclude sql.NullString
		var hoursOld, resultsWanted sql.NullInt64
		var lastRun sql.NullTime

		err := rows.Scan(&s.ID, &s.Keyword, &s.Country, &location, &language, &s.Frequency, &hoursOld, &exclude, &resultsWanted, &lastRun)
		if err != nil {
			continue
		}

		s.UserID = userID
		if location.Valid {
			s.Location = location.String
		}
		if language.Valid {
			s.Language = language.String
		}
		if hoursOld.Valid {
			s.HoursOld = int(hoursOld.Int64)
		}
		if exclude.Valid {
			s.Exclude = exclude.String
		}
		if resultsWanted.Valid {
			s.ResultsWanted = int(resultsWanted.Int64)
		}
		if lastRun.Valid {
			s.LastRun = lastRun.Time
		}

		searches = append(searches, s)
	}

	// Return empty array if no searches
	if searches == nil {
		searches = []models.UserSearch{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searches)
}

func createSearchHandler(w http.ResponseWriter, r *http.Request) {

	// 1. Verify Auth Token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, "Invalid token email", http.StatusUnauthorized)
		return
	}

	// Get User ID and Subscription
	var userID int
	var subscription string
	err = db.DB.QueryRow("SELECT id, subscription_plan FROM users WHERE email = ?", email).Scan(&userID, &subscription)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check subscription - only pro users can save alerts
	if subscription != "pro" {
		http.Error(w, "This feature requires a Pro subscription", http.StatusForbidden)
		return
	}

	// Parse Request
	var req models.CreateSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default frequency
	if req.Frequency == "" {
		req.Frequency = "hourly"
	}

	// Check if search already exists
	var existingID int
	err = db.DB.QueryRow(`
		SELECT id FROM user_searches 
		WHERE user_id = ? AND keyword = ? AND country = ? AND location = ? AND language = ? AND hours_old = ? AND exclude = ? AND results_wanted = ?
	`, userID, req.Keyword, req.Country, req.Location, req.Language, req.HoursOld, req.Exclude, req.ResultsWanted).Scan(&existingID)

	if err == nil {
		// Search already exists
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Alert already exists",
			"id":      existingID,
		})
		return
	} else if err != sql.ErrNoRows {
		// Database error
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Insert Search (only if it doesn't exist)
	_, err = db.DB.Exec(`
        INSERT INTO user_searches (user_id, keyword, country, location, language, frequency, hours_old, exclude, results_wanted, last_run)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)
    `, userID, req.Keyword, req.Country, req.Location, req.Language, req.Frequency, req.HoursOld, req.Exclude, req.ResultsWanted)

	if err != nil {
		http.Error(w, "Failed to save search: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Search saved successfully"})
}

func deleteSearchHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Verify Auth Token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, "Invalid token email", http.StatusUnauthorized)
		return
	}

	// Get User ID
	var userID int
	err = db.DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Extract search ID from URL path (e.g., /api/searches/123)
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	searchIDStr := pathParts[len(pathParts)-1]
	searchID, err := strconv.Atoi(searchIDStr)
	if err != nil {
		http.Error(w, "Invalid search ID", http.StatusBadRequest)
		return
	}

	// Log the deletion attempt
	log.Printf("[Delete Alert] User %d attempting to delete search ID: %d", userID, searchID)

	// Delete the search, but only if it belongs to this user
	result, err := db.DB.Exec("DELETE FROM user_searches WHERE id = ? AND user_id = ?", searchID, userID)
	if err != nil {
		log.Printf("[Delete Alert] Error deleting search ID %d: %v", searchID, err)
		http.Error(w, "Failed to delete search", http.StatusInternalServerError)
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[Delete Alert] Error getting rows affected for search ID %d: %v", searchID, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("[Delete Alert] Rows affected: %d", rowsAffected)

	if rowsAffected == 0 {
		log.Printf("[Delete Alert] Search ID %d not found or unauthorized for user %d", searchID, userID)
		http.Error(w, "Search not found or unauthorized", http.StatusNotFound)
		return
	}

	log.Printf("[Delete Alert] Successfully deleted search ID %d for user %d", searchID, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Search deleted successfully"})
}
