package auth

import (
	"go.uber.org/zap"
	"net/http"
)

func (a *Autentificator) MiddlewareCookies(h http.HandlerFunc) http.HandlerFunc {
	a.Log.Debug("Проверяем куки")
	var id int
	auth := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		if err != nil {
			w = a.setCookies(w)
			a.Log.Debug("Устанавливаем куки")

		}

		id = a.readToken(cookie.Value)
		a.Log.Debug("Устанавливаем в контекст", zap.Int("id", id))

		r = r.WithContext(a.setContext(r.Context(), id))
		h.ServeHTTP(w, r)
	}
	return auth
}
