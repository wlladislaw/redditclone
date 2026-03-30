package middleware

import (
	"context"
	"fmt"
	"net/http"
	"redditclone/pkg/session"
	"strings"
)

func AuthMiddleware(sm *session.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := GetTokenFromHeader(r)
			if !ok {
				http.Error(w, "no auth session in header", http.StatusUnauthorized)
				return
			}
			sess, err := sm.Check(token)
			if err != nil {
				fmt.Printf("no auth: %v", err)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), session.SessKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetTokenFromHeader(req *http.Request) (string, bool) {
	headerBearer := req.Header.Get("Authorization")
	parts := strings.Fields(headerBearer)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}
	return parts[1], true
}
