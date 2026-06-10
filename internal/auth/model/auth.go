package model

import "time"

type TokenWithUserEmailToRefreshToken struct {
	ID        int
	UserID    int
	Email     string
	Username  string
	TokenHash string
	ExpiresAt time.Time
}
