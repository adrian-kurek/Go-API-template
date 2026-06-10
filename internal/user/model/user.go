package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID            int
	Email         string
	Username      string
	Password      string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserClaims struct {
	ID       int    `json:"id" example:"11"`
	Email    string `json:"email" example:"joedoe@email.com"`
	Username string `json:"username" example:"slodkiadrianek"`
	Exp      int64
	jwt.RegisteredClaims
}
