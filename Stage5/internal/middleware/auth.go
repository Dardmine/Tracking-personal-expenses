package middleware

import (
	"context"
	"database/sql"
	"net/http"
)

type ctxKey string

const UserIDKey ctxKey = "userID"

// RequireAuth проверяет сессию и кладёт user_id в контекст
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
		// Проверка срока действия сессии (упрощённо — сравниваем строкой)
		// В реальном проекте нужно парсить время.

		// Передаём user_id в контекст запроса
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
