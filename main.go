package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"jobseek-web-be/internal/db"
	"jobseek-web-be/internal/handlers"
	"jobseek-web-be/internal/scheduler"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

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
	// Protected Routes
	http.HandleFunc("/api/search", handlers.SearchHandler)
	http.HandleFunc("/api/redirect", handlers.RedirectHandler)
	http.HandleFunc("/api/unsubscribe", handlers.UnsubscribeHandler)
	http.HandleFunc("/api/searches", handlers.SaveSearchHandler)

	// Start Scheduler
	scheduler := scheduler.NewScheduler()
	scheduler.Start()
	defer scheduler.Stop()

	// Serve Static Files (Frontend)
	// Serve Static Files (Frontend) with SPA support
	frontendFS := http.FileServer(http.Dir("../jobseek-web-fe/dist"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If path starts with /api, let it fall through (though /api is handled by specific routes above)
		// But since we use exact matches or prefixes, explicit API routes take precedence.

		// Check if file exists in dist
		path := "../jobseek-web-fe/dist" + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// File does not exist, serve index.html
			http.ServeFile(w, r, "../jobseek-web-fe/dist/index.html")
			return
		}
		// Serves the actual file
		frontendFS.ServeHTTP(w, r)
	})

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "jobseek-web-be"})
}
