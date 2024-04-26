package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims //nolint: typecheck
	UserID               int
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "ty89huj9j"
const SCHEME = "Bearer "
