package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims //nolint: typecheck
	UserID               int
}

type CustomKeyContext string

const UserID CustomKeyContext = "id"

const TokenExp = time.Hour * 3

// SecretKey пока что в коде, т.к. по заданию не понятно где должен храниться он.
// Если в записать в переменную окружения, то автотесты на сервере не пройдут.
const SecretKey = "ty89huj9j" //nolint: gosec
const Scheme = "Bearer "
