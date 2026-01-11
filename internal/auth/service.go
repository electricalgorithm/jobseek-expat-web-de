package auth

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"jobseek-web-be/internal/db"
	"jobseek-web-be/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var SecretKey = []byte("super-secret-key-change-this-in-prod")

func RegisterUser(req models.RegisterRequest) error {
	// 1. Check if user exists
	var exists bool
	err := db.DB.QueryRow("SELECT exists(SELECT 1 FROM users WHERE email=?)", req.Email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if exists {
		return errors.New("user already exists")
	}

	// 2. Mock payment verification for 'pro' plan
	if req.Subscription == "pro" {
		if req.PaymentToken == "" {
			return errors.New("payment token required for pro plan")
		}
		// In a real app, verify Stripe token here
		log.Printf("Verifying mocked payment token %s for user %s", req.PaymentToken, req.Email)
	} else {
		req.Subscription = "basic"
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 4. Calculate trial end date (7 days from now)
	trialEndsAt := time.Now().Add(7 * 24 * time.Hour)

	// 5. Insert user with trial period
	_, err = db.DB.Exec(
		"INSERT INTO users(name, email, password, subscription_plan, trial_ends_at) VALUES(?, ?, ?, ?, ?)",
		req.Name, req.Email, string(hashedPassword), req.Subscription, trialEndsAt,
	)
	return err
}

func LoginUser(creds models.Credentials) (string, string, string, error) {
	var storedPassword string
	var subscription string
	var name string
	var trialEndsAt sql.NullTime

	err := db.DB.QueryRow(
		"SELECT password, subscription_plan, name, trial_ends_at FROM users WHERE email=?",
		creds.Email,
	).Scan(&storedPassword, &subscription, &name, &trialEndsAt)

	if err == sql.ErrNoRows {
		return "", "", "", errors.New("Invalid credentials")
	} else if err != nil {
		return "", "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)); err != nil {
		return "", "", "", errors.New("Invalid credentials")
	}

	// Check if trial has expired
	if trialEndsAt.Valid && time.Now().After(trialEndsAt.Time) {
		return "", "", "", errors.New("Trial period has expired - please upgrade your subscription")
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"email": creds.Email,
		"name":  name,
		"sub":   subscription,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}

	// Add trial info if exists
	if trialEndsAt.Valid {
		claims["trial_ends_at"] = trialEndsAt.Time.Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SecretKey)
	return tokenString, subscription, name, err
}
