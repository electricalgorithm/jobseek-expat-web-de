package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"jobseek-web-be/internal/auth"
	"jobseek-web-be/internal/db"
	"jobseek-web-be/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

func SaveSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
		WHERE user_id = ? AND keyword = ? AND country = ? AND location = ? AND language = ?
	`, userID, req.Keyword, req.Country, req.Location, req.Language).Scan(&existingID)

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
        INSERT INTO user_searches (user_id, keyword, country, location, language, frequency, last_run)
        VALUES (?, ?, ?, ?, ?, ?, NULL)
    `, userID, req.Keyword, req.Country, req.Location, req.Language, req.Frequency)

	if err != nil {
		http.Error(w, "Failed to save search: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Search saved successfully"})
}
