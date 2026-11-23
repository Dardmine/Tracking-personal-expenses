package middleware

import (
	"context"
	"database/sql"
	"net/http"
)

type ctxKey string

const UserIDKey ctxKey = "userID"

func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var userID int
		var expiresAt string
		err = db.QueryRow(
			"SELECT user_id, expires_at FROM sessions WHERE session_id = ?",
			cookie.Value,
		).Scan(&userID, &expiresAt)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
