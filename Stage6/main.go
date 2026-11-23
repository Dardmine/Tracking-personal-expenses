package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	_ "modernc.org/sqlite"

	"Stage6/internal/handlers"
	"Stage6/internal/middleware"
	"Stage6/internal/repository"
)

func main() {
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatal("sql.Open:", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatal("миграции:", err)
	}

	expenseRepo := repository.NewExpenseRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	userRepo := repository.NewUserRepository(db)

	expenseHandler := &handlers.ExpenseHandler{
		Repo:    expenseRepo,
		CatRepo: categoryRepo,
	}
	authHandler := &handlers.AuthHandler{
		Users: userRepo,
		DB:    db,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)

	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.RequireAuth(db, h)
	}

	mux.HandleFunc("/expenses", protected(expenseHandler.List))
	mux.HandleFunc("/expenses/add", protected(expenseHandler.Add))
	mux.HandleFunc("/expenses/edit", protected(expenseHandler.EditForm))
	mux.HandleFunc("/expenses/update", protected(expenseHandler.Update))
	mux.HandleFunc("/expenses/delete", protected(expenseHandler.Delete))

	statsRepo := repository.NewStatsRepository(db)
	statsHandler := &handlers.StatsHandler{Repo: statsRepo}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/stats", protected(statsHandler.Page))
	mux.HandleFunc("/api/stats/by-category", protected(statsHandler.ByCategory))
	mux.HandleFunc("/api/stats/by-month", protected(statsHandler.ByMonth))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
	})

	log.Println("Сервер запущен на http://localhost:8080")
	log.Println("Регистрация: http://localhost:8080/register")
	log.Println("Вход:        http://localhost:8080/login")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func runMigrations(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS expenses (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		amount      REAL NOT NULL,
		description TEXT NOT NULL,
		date        TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS categories (
		id   INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT OR IGNORE INTO categories (name) VALUES
		('Еда'), ('Транспорт'), ('Жильё'), ('Развлечения'), ('Прочее')`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`ALTER TABLE expenses ADD COLUMN category_id INTEGER REFERENCES categories(id)`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		email         TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		session_id TEXT PRIMARY KEY,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		expires_at DATETIME NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`ALTER TABLE expenses ADD COLUMN user_id INTEGER REFERENCES users(id)`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	return nil
}
