package scheduler

import (
	"database/sql"
	"log"
	"os"
	"time"

	"jobseek-web-be/internal/db"
	"jobseek-web-be/internal/email"
	"jobseek-web-be/internal/search"

	"github.com/robfig/cron/v3"
)

type JobScheduler struct {
	cron *cron.Cron
}

func NewScheduler() *JobScheduler {
	return &JobScheduler{
		cron: cron.New(),
	}
}

func (s *JobScheduler) Start() {
	// Schedule job to run periodically
	// Default: "@hourly"
	freq := os.Getenv("SCHEDULER_FREQUENCY")
	if freq == "" {
		freq = "@hourly"
	}

	_, err := s.cron.AddFunc(freq, func() {
		log.Printf("[Scheduler] Starting job search task (Schedule: %s)...", freq)
		RunJobSearchTask()
	})

	if err != nil {
		log.Fatalf("Error scheduling job: %v", err)
	}

	s.cron.Start()
	log.Printf("Scheduler started. Jobs running with frequency: %s", freq)
}

func (s *JobScheduler) Stop() {
	s.cron.Stop()
}

func RunJobSearchTask() {
	// 1. Fetch all active searches into memory to avoid locking the DB during long processing
	type SearchTask struct {
		ID            int
		UserID        int
		Keyword       string
		Country       string
		Location      string
		Language      string
		UserEmail     string
		UserName      string
		Frequency     string
		HoursOld      sql.NullInt64
		Exclude       sql.NullString
		ResultsWanted sql.NullInt64
		LastRun       sql.NullTime
	}

	var tasks []SearchTask

	rows, err := db.DB.Query(`
		SELECT us.id, us.user_id, us.keyword, us.country, us.location, us.language, u.email, u.name, us.frequency, us.hours_old, us.exclude, us.results_wanted, us.last_run
		FROM user_searches us 
		JOIN users u ON us.user_id = u.id
	`)
	if err != nil {
		log.Printf("[Scheduler] Error fetching searches: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t SearchTask
		var loc, lang sql.NullString

		if err := rows.Scan(&t.ID, &t.UserID, &t.Keyword, &t.Country, &loc, &lang, &t.UserEmail, &t.UserName, &t.Frequency, &t.HoursOld, &t.Exclude, &t.ResultsWanted, &t.LastRun); err != nil {
			log.Printf("[Scheduler] Error scanning row: %v", err)
			continue
		}

		if loc.Valid {
			t.Location = loc.String
		}
		if lang.Valid {
			t.Language = lang.String
		}

		tasks = append(tasks, t)
	}
	rows.Close() // Explicitly close before processing

	// 2. Process tasks
	for _, t := range tasks {
		// Check frequency
		if t.LastRun.Valid {
			nextRun := t.LastRun.Time
			switch t.Frequency {
			case "hourly":
				nextRun = nextRun.Add(1 * time.Hour)
			case "daily":
				nextRun = nextRun.Add(24 * time.Hour)
			default:
				nextRun = nextRun.Add(1 * time.Hour)
			}

			if time.Now().Before(nextRun) {
				continue
			}
		}

		log.Printf("[Scheduler] Processing alert for user %s: %s in %s", t.UserEmail, t.Keyword, t.Country)

		// Create SearchParams from task
		hoursOld := 24 // Default
		if t.HoursOld.Valid {
			hoursOld = int(t.HoursOld.Int64)
		}

		exclude := ""
		if t.Exclude.Valid {
			exclude = t.Exclude.String
		}

		resultsWanted := 10 // Default
		if t.ResultsWanted.Valid {
			resultsWanted = int(t.ResultsWanted.Int64)
		}

		// Execute Search
		params := search.SearchParams{
			Keyword:       t.Keyword,
			Country:       t.Country,
			Location:      t.Location,
			LocalLanguage: t.Language,
			ResultsWanted: resultsWanted,
			HoursOld:      hoursOld,
			Exclude:       exclude,
		}

		results, err := search.ExecuteSearch(params)
		if err != nil {
			log.Printf("[Scheduler] Search failed for %d: %v", t.ID, err)
			continue
		}

		if len(results) == 0 {
			log.Printf("[Scheduler] No results found for %d", t.ID)
			continue
		}

		// Send Email
		if err := email.SendJobAlert(t.UserEmail, t.UserName, t.UserID, t.ID, results); err != nil {
			log.Printf("[Scheduler] Failed to send email to %s: %v", t.UserEmail, err)
		}

		// Update Last Run
		_, err = db.DB.Exec("UPDATE user_searches SET last_run = ? WHERE id = ?", time.Now(), t.ID)
		if err != nil {
			log.Printf("[Scheduler] Failed to update last_run for %d: %v", t.ID, err)
		}
	}
}
