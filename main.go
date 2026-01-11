package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"jobseek-web-be/internal/db"
	"jobseek-web-be/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Database
	db.InitDB()

	// API Routes
	http.HandleFunc("/api/health", healthHandler)

	// Public Auth Routes
	http.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	http.HandleFunc("/api/auth/login", handlers.LoginHandler)

	// Protected Routes (example usage)
	// For now, we keep search public as the "Basic" plan functionality implies access.
	// In strict mode, we would wrap it: http.HandleFunc("/api/search", middleware.AuthMiddleware(handlers.SearchHandler))
	http.HandleFunc("/api/search", handlers.SearchHandler)

	// FUTURE: Cron job scheduler for 'pro' users would go here
	// go startScheduler()

	// Serve Static Files (Frontend)
	fs := http.FileServer(http.Dir("../jobseek-web-fe/dist"))
	http.Handle("/", fs)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "jobseek-web-be"})
}
