package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	_ "modernc.org/sqlite"

	"Stage5/internal/handlers"
	"Stage5/internal/middleware"
	"Stage5/internal/repository"
)

func main() {
	// 1. Открываем базу данных (файл создастся автоматически)
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatal("sql.Open:", err)
	}
	defer db.Close()

	// 2. Миграции. Выполняем по порядку; повторный запуск не ломает схему.
	if err := runMigrations(db); err != nil {
		log.Fatal("миграции:", err)
	}

	// 3. Репозитории
	expenseRepo := repository.NewExpenseRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 4. Обработчики
	expenseHandler := &handlers.ExpenseHandler{
		Repo:    expenseRepo,
		CatRepo: categoryRepo,
	}
	authHandler := &handlers.AuthHandler{
		Users: userRepo,
		DB:    db,
	}

	// 5. Маршруты
	mux := http.NewServeMux()

	// --- Публичные маршруты (без авторизации) ---
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)

	// --- Защищённые маршруты (только для авторизованных) ---
	// Оборачиваем обработчики в middleware.RequireAuth.
	// Цепочка: RequireAuth → сам обработчик.
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.RequireAuth(db, h)
	}

	mux.HandleFunc("/expenses", protected(expenseHandler.List))
	mux.HandleFunc("/expenses/add", protected(expenseHandler.Add))
	mux.HandleFunc("/expenses/edit", protected(expenseHandler.EditForm))
	mux.HandleFunc("/expenses/update", protected(expenseHandler.Update))
	mux.HandleFunc("/expenses/delete", protected(expenseHandler.Delete))

	// --- Корень: перенаправляем на список трат ---
	// Если пользователь не авторизован, RequireAuth сам отправит на /login.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
	})

	// 6. Запуск сервера
	log.Println("Сервер запущен на http://localhost:8080")
	log.Println("Регистрация: http://localhost:8080/register")
	log.Println("Вход:        http://localhost:8080/login")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// runMigrations создаёт таблицы и добавляет недостающие колонки.
// Все операции идемпотентны: повторный запуск не ломает схему.
func runMigrations(db *sql.DB) error {
	// --- Этап 3: таблица expenses ---
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS expenses (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		amount      REAL NOT NULL,
		description TEXT NOT NULL,
		date        TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	// --- Этап 4: категории и поле category_id ---
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

	// ALTER TABLE падает, если колонка уже есть — игнорируем эту конкретную ошибку.
	_, err = db.Exec(`ALTER TABLE expenses ADD COLUMN category_id INTEGER REFERENCES categories(id)`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	// --- Этап 5: пользователи, сессии, user_id в expenses ---
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
