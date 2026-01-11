package models

import "time"

type User struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Password         string    `json:"password"`
	SubscriptionPlan string    `json:"subscription_plan"`
	CreatedAt        time.Time `json:"created_at"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Subscription string `json:"subscription"`  // "basic" or "pro"
	PaymentToken string `json:"payment_token"` // Mock token
}

type AuthResponse struct {
	Token        string `json:"token"`
	Subscription string `json:"subscription"`
	Name         string `json:"name"`
	Email        string `json:"email"`
}
