package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"Stage5/internal/repository"
)

type AuthHandler struct {
	Users *repository.UserRepository
	DB    *sql.DB
}

// Регистрация
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "layout_register", nil)
		return
	}
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Хэшируем пароль через bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хэширования", http.StatusInternalServerError)
		return
	}

	if err := h.Users.Create(email, string(hash)); err != nil {
		http.Error(w, "Пользователь уже существует", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Вход
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "layout_login", nil)
		return
	}
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.Users.GetByEmail(email)
	if err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// Сравниваем введённый пароль с хэшем
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// Генерируем случайный session_id
	b := make([]byte, 32)
	rand.Read(b)
	sessionID := hex.EncodeToString(b)

	// Сохраняем сессию в БД
	_, err = h.DB.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, user.ID, time.Now().Add(24*time.Hour),
	)
	if err != nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	// Ставим cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// Выход
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		h.DB.Exec("DELETE FROM sessions WHERE session_id = ?", cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
