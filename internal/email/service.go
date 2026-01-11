package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/resend/resend-go/v3"
)

type JobResult struct {
	Title   string `json:"title"`
	Company string `json:"company"`
	Url     string `json:"job_url"`
}

type EmailData struct {
	AppName        string
	UserName       string
	UserID         int
	SearchID       int
	JobCount       int
	Jobs           []JobResult
	MoreCount      int
	UnsubscribeURL string
}

func SendJobAlert(toEmail, userName string, userID, searchID int, jobs []interface{}) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "JobSeek"
	}

	domain := os.Getenv("APP_DOMAIN")
	if domain == "" {
		domain = "http://localhost:8080"
	}

	// Prepare Data
	var jobList []JobResult
	for i, job := range jobs {
		if i >= 10 {
			break
		}
		j, ok := job.(map[string]interface{})
		if ok {
			title := j["title"].(string)
			company := j["company"].(string)
			rawUrl := j["job_url"].(string)
			if title == "" {
				title = "Job Opening"
			}
			if company == "" {
				company = "Unknown Company"
			}

			// Wrap URL with Redirect
			encodedUrl := base64.URLEncoding.EncodeToString([]byte(rawUrl))
			redirectUrl := fmt.Sprintf("%s/api/redirect?data=%s", domain, encodedUrl)

			jobList = append(jobList, JobResult{
				Title:   title,
				Company: company,
				Url:     redirectUrl,
			})
		}
	}

	moreCount := 0
	if len(jobs) > 10 {
		moreCount = len(jobs) - 10
	}

	unsubscribeURL := fmt.Sprintf("%s/unsubscribe?uid=%d&sid=%d", domain, userID, searchID)

	data := EmailData{
		AppName:        appName,
		UserName:       userName,
		UserID:         userID,
		SearchID:       searchID,
		JobCount:       len(jobs),
		Jobs:           jobList,
		MoreCount:      moreCount,
		UnsubscribeURL: unsubscribeURL,
	}

	// Mock Fallback
	if apiKey == "" {
		log.Println("[Email] RESEND_API_KEY is missing. Falling back to mock email.")
		return mockSend(toEmail, userName, jobs)
	}

	client := resend.NewClient(apiKey)
	subject := fmt.Sprintf("Found %d New Jobs For You!", len(jobs))

	htmlContent, err := renderTemplate(data)
	if err != nil {
		log.Printf("[Email] Failed to render template: %v", err)
		return err
	}

	// Construct Sender Name
	fromName := fmt.Sprintf("%s Expat", appName)
	// fromEmail := fmt.Sprintf("jobs@%s.com", appName)

	// For Resend specifically, domains must be verified. We'll use the one from env or safe default.
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = fmt.Sprintf("%s <jobs@expatter.gyokhan.com>", fromName)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlContent,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("[Email] Failed to send email via Resend: %v", err)
		return err
	}

	log.Printf("[Email] Sent email to %s via Resend. ID: %s", toEmail, sent.Id)
	return nil
}

func renderTemplate(data EmailData) (string, error) {
	// Locate template file.
	// In production, this path needs to relate to the binary location or be embedded.
	// For dev, we look in internal/email/template.html
	cwd, _ := os.Getwd()
	tmplPath := filepath.Join(cwd, "internal", "email", "template.html")

	// Fallback check if running from nested dir or bin?
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		// Try root relative?
		tmplPath = "internal/email/template.html"
	}

	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return buf.String(), nil
}

func mockSend(toEmail, userName string, jobs []interface{}) error {
	log.Printf("---------------------------------------------------")
	log.Printf("MOCK EMAIL TO: %s", toEmail)
	log.Printf("SUBJECT: New Job Matches Found for %s!", userName)
	log.Printf("BODY:")
	log.Printf("Hi %s,\n\nWe found %d new jobs matching your criteria:\n", userName, len(jobs))

	count := 0
	for _, job := range jobs {
		if count >= 5 {
			log.Printf("... and %d more.", len(jobs)-5)
			break
		}
		j, ok := job.(map[string]interface{})
		if ok {
			log.Printf("- %s at %s: %s", j["title"], j["company"], j["job_url"])
		}
		count++
	}
	log.Printf("---------------------------------------------------")
	return nil
}
