package auth

import (
	"go.uber.org/zap"
	"net/http"
)

func (a *Autentificator) MiddlewareCookies(h http.HandlerFunc) http.HandlerFunc {
	auth := func(w http.ResponseWriter, r *http.Request) {
		var id int
		ctx := r.Context()

		header := r.Header.Get("Autorization")
		if header != "" {
			a.Log.Debug("Достаем ID из хеадера")
			id = a.readToken(ctx, header)
			if id == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			a.Log.Debug("Устанавливаем в контекст", zap.Int("id", id))
			r = r.WithContext(a.setContext(r.Context(), id))
			h.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("Authorization")

		if err != nil {
			a.Log.Debug("Устанавливаем куки")
			w, id = a.setCookies(ctx, w)
		} else {
			a.Log.Debug("Достаем ID из куки")
			id = a.readToken(ctx, cookie.Value)
			if id == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		a.Log.Debug("Устанавливаем в контекст", zap.Int("id", id))
		r = r.WithContext(a.setContext(r.Context(), id))
		h.ServeHTTP(w, r)
	}
	return auth
}
