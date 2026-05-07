package httpapi

import (
	"context"
	"net/http"
	"strings"

	"devops-pipeline/internal/auth"
	"devops-pipeline/internal/store"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
)

func AuthMiddleware(jwtManager *auth.JWTManager, tokenStore *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := bearerTokenFromHeader(r)
			if !ok {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			userID, username, err := authenticateToken(r.Context(), jwtManager, tokenStore, tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// 将用户信息存入context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UsernameKey, username)

			// 继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerTokenFromHeader(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return "", false
	}

	return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[1]) != ""
}

func authenticateToken(ctx context.Context, jwtManager *auth.JWTManager, tokenStore *store.Store, tokenString string) (int64, string, error) {
	claims, err := jwtManager.ValidateToken(tokenString)
	if err == nil {
		return claims.UserID, claims.Username, nil
	}

	apiToken, tokenErr := tokenStore.ValidateAPIToken(ctx, tokenString)
	if tokenErr == nil {
		return 0, "api-token:" + apiToken.Name, nil
	}

	return 0, "", err
}

func GetUserID(r *http.Request) int64 {
	if userID, ok := r.Context().Value(UserIDKey).(int64); ok {
		return userID
	}
	return 0
}

func GetUsername(r *http.Request) string {
	if username, ok := r.Context().Value(UsernameKey).(string); ok {
		return username
	}
	return ""
}
