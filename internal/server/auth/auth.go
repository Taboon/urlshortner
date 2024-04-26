package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/Taboon/urlshortner/internal/storage"
	jwt "github.com/golang-jwt/jwt/v4"
)

type Autentificator struct {
	Log *zap.Logger
	R   storage.Repository
}

func NewAuthentificator(l *zap.Logger, r storage.Repository) Autentificator {
	return Autentificator{
		Log: l,
		R:   r,
	}
}

func (a *Autentificator) readToken(token string) int {
	token = strings.TrimPrefix(token, SCHEME)
	return a.getUserID(token)
}

func (a *Autentificator) getUserID(token string) int {
	a.Log.Debug("Получаем из токена userID")
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID
}

func (a *Autentificator) setContext(ctx context.Context, id int) context.Context {
	a.Log.Debug("Устанавливаем контекст")
	return context.WithValue(ctx, "id", id)
}

func (a *Autentificator) setCookies(w http.ResponseWriter) http.ResponseWriter {
	cookie, err := a.signCookies()
	if err != nil {
		a.Log.Error("Ошибка установки куков", zap.Error(err))
		return w
	}
	http.SetCookie(w, cookie)
	return w
}

func (a *Autentificator) signCookies() (*http.Cookie, error) {

	token, err := a.buildJWTString()
	if err != nil {
		return nil, err
	}
	a.Log.Debug("Подписываем куки", zap.String("token", token))
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    fmt.Sprintf("%v%v", SCHEME, token),
		Secure:   false,
		HttpOnly: true,
		SameSite: 1,
	}
	return &cookie, err
}

func (a *Autentificator) buildJWTString() (string, error) {
	a.Log.Debug("Получаем закодированный токен")
	id, err := a.getNewUserID()
	if err != nil {
		return "", err
	}
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: id,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (a *Autentificator) getNewUserID() (int, error) {
	c := 100
	b := make([]byte, c)
	id, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return 0, err
	}
	a.Log.Debug("Получаем новый userID", zap.Int("id", id))
	return id, nil
}
