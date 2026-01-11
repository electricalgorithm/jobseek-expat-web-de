package handlers

import (
	"encoding/json"
	"jobseek-web-be/internal/db"
	"log"
	"net/http"
)

type UnsubscribeRequest struct {
	UserID         int  `json:"user_id"`
	SearchID       *int `json:"search_id"` // Optional
	UnsubscribeAll bool `json:"unsubscribe_all"`
}

func UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UnsubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.UnsubscribeAll {
		// Unsubscribe all searches for the user
		_, err := db.DB.Exec("DELETE FROM user_searches WHERE user_id = ?", req.UserID)
		if err != nil {
			log.Printf("Unsubscribe all failed for user %d: %v", req.UserID, err)
			http.Error(w, "Failed to unsubscribe", http.StatusInternalServerError)
			return
		}
		log.Printf("Unsubscribed all searches for user %d", req.UserID)
	} else if req.SearchID != nil {
		// Unsubscribe specific search
		// Verify ownership first for security (simple check)
		_, err := db.DB.Exec("DELETE FROM user_searches WHERE id = ? AND user_id = ?", *req.SearchID, req.UserID)
		if err != nil {
			log.Printf("Unsubscribe search %d failed: %v", *req.SearchID, err)
			http.Error(w, "Failed to unsubscribe", http.StatusInternalServerError)
			return
		}
		log.Printf("Unsubscribed search %d for user %d", *req.SearchID, req.UserID)
	} else {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
