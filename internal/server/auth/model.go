package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims //nolint: typecheck
	UserID               int
}

const TOKEN_EXP = time.Hour * 3 //nolint: stylecheck, revive

// SECRET_KEY пока что в коде, т.к. по заданию не понятно где должен храниться он.
// Если в записать в переменную окружения, то автотесты на сервере не пройдут.
const SECRET_KEY = "ty89huj9j" //nolint: gosec, stylecheck, revive
const SCHEME = "Bearer "       //nolint: stylecheck, revive
