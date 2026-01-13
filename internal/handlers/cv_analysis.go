package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"jobseek-web-be/internal/auth"
	"jobseek-web-be/internal/db"
	"jobseek-web-be/internal/search"

	"github.com/golang-jwt/jwt/v5"
)

// CVAnalysisRequest represents the request parameters for CV analysis
type CVAnalysisRequest struct {
	Countries string `json:"countries"` // Comma-separated countries
	Locations string `json:"locations"` // Comma-separated locations
}

// CVAnalysisResponse represents the analysis result from jobseek-expat
type CVAnalysisResponse struct {
	Success  bool                   `json:"success"`
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}

// AnalyzeCVHandler handles CV upload and analysis (Pro users only)
func AnalyzeCVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Verify Auth Token and Pro status
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

	// Get user from database
	var userID int
	var subscription string
	var paid int
	var createdAt time.Time

	err = db.DB.QueryRow("SELECT id, subscription_plan, paid, created_at FROM users WHERE email = ?", email).
		Scan(&userID, &subscription, &paid, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if user is Pro or in trial
	trialDays := 7
	daysSinceRegistration := int(time.Since(createdAt).Hours() / 24)
	isPro := subscription == "pro" && paid == 1
	isInTrial := daysSinceRegistration < trialDays

	if !isPro && !isInTrial {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "CV analysis is a Pro feature. Please upgrade your subscription.",
		})
		return
	}

	// 2. Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("cv")
	if err != nil {
		http.Error(w, "Missing CV file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".docx" && ext != ".txt" {
		http.Error(w, "Invalid file format. Supported: PDF, DOCX, TXT", http.StatusBadRequest)
		return
	}

	// Get countries and locations from form (optional)
	countries := r.FormValue("countries")
	if countries == "" {
		countries = "Germany" // Default
	}

	locations := r.FormValue("locations")
	if locations == "" {
		locations = "Remote" // Default
	}

	// 3. Save to temporary file
	tempFile, err := os.CreateTemp("", fmt.Sprintf("cv_*%s", ext))
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up

	_, err = io.Copy(tempFile, file)
	tempFile.Close()
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// 4. Call jobseek-expat analyze-cv command
	cmdPath := search.GetJobSeekPath()
	if cmdPath == "" {
		http.Error(w, "jobseek-expat CLI not found", http.StatusInternalServerError)
		return
	}

	args := []string{
		"analyze-cv",
		tempPath,
		"--countries", countries,
		"--locations", locations,
		"--output", "json",
	}

	// Check if GEMINI_API_KEY is set
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		http.Error(w, "GEMINI_API_KEY environment variable not set. Please configure it to use CV analysis.", http.StatusInternalServerError)
		return
	}

	// Set GEMINI_API_KEY from environment
	cmd := exec.Command(cmdPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GEMINI_API_KEY=%s", geminiKey))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("CV Analysis failed: %s\nOutput: %s", err, string(output))
		http.Error(w, fmt.Sprintf("Analysis failed: %s", string(output)), http.StatusInternalServerError)
		return
	}

	// 5. Parse JSON response
	var analysisResult CVAnalysisResponse
	err = json.Unmarshal(output, &analysisResult)
	if err != nil {
		http.Error(w, "Failed to parse analysis result", http.StatusInternalServerError)
		return
	}

	// 6. Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysisResult)
}
