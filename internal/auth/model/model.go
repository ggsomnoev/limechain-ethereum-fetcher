package model

import "time"

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type Token struct {
	Value     string    `json:"value"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expiresAt"`
}
