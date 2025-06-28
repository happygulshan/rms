package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"rms/jwt_utils"
	"strings"
)

type key string

const userIDKey key = "userID"
const rolesKey key = "rolesID"

func AuthMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := strings.TrimSpace(r.Header.Get("Authorization"))
			if token == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			// userID, err := jwt_utils.ValidateJWT(token)

			userID, roles, err := jwt_utils.ValidateAccessJWT(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, rolesKey, roles)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

// Helper to get userID from context
func GetUserID(r *http.Request) string {
	id := r.Context().Value(userIDKey)
	if idStr, ok := id.(string); ok {
		return idStr
	}
	return ""
}

func GetUserRoles(r *http.Request) []string {
	roles := r.Context().Value(rolesKey)
	if roles, ok := roles.([]string); ok {
		return roles
	}
	return nil
}
