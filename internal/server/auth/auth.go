package auth

import (
	"context"
	"fmt"
	"github.com/Taboon/urlshortner/internal/config"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/Taboon/urlshortner/internal/storage"
	jwt "github.com/golang-jwt/jwt/v4"
)

type Autentificator struct {
	Log     *zap.Logger
	R       storage.Repository
	BaseURL config.Address
}

func NewAuthentificator(l *zap.Logger, r storage.Repository, bu config.Address) Autentificator {
	return Autentificator{
		Log:     l,
		R:       r,
		BaseURL: bu,
	}
}

func (a *Autentificator) readToken(ctx context.Context, token string) int {
	token = strings.TrimPrefix(token, Scheme)
	return a.getUserID(ctx, token)
}

func (a *Autentificator) getUserID(_ context.Context, token string) int {
	a.Log.Debug("Получаем из токена userID")
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	_, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		a.Log.Error("Ошибка парсинга ID", zap.Error(err))
		return 0
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID
}

func (a *Autentificator) setContext(ctx context.Context, id int) context.Context {
	a.Log.Debug("Устанавливаем контекст")
	return context.WithValue(ctx, storage.UserID, id) //nolint: revive, staticcheck
}

func (a *Autentificator) setCookies(ctx context.Context, w http.ResponseWriter) (http.ResponseWriter, int) {
	cookie, id, err := a.SignCookies(ctx)
	if err != nil {
		a.Log.Error("Ошибка установки куков", zap.Error(err))
		return w, 0
	}
	http.SetCookie(w, cookie)
	a.Log.Debug("Установили куки", zap.Any("cookie", cookie))
	return w, id
}

func (a *Autentificator) SignCookies(ctx context.Context) (*http.Cookie, int, error) {
	token, id, err := a.buildJWTString(ctx)
	if err != nil {
		return nil, 0, err
	}
	a.Log.Debug("Подписываем куки", zap.String("token", token))
	cookie := http.Cookie{
		Name:  "Authorization",
		Value: fmt.Sprintf("%v%v", Scheme, token),
		//Secure:   true,
		//HttpOnly: true,
		MaxAge: 3600,
		//Domain:   a.BaseURL.String(),
		SameSite: http.SameSiteNoneMode,
	}
	return &cookie, id, err
}

func (a *Autentificator) buildJWTString(ctx context.Context) (string, int, error) {
	a.Log.Debug("Получаем закодированный токен")
	id, err := a.getNewUserID(ctx)
	if err != nil {
		return "", 0, err
	}
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// собственное утверждение
		UserID: id,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", 0, err
	}

	// возвращаем строку токена
	return tokenString, id, nil
}

func (a *Autentificator) getNewUserID(ctx context.Context) (int, error) {
	return a.R.GetNewUser(ctx)
}
