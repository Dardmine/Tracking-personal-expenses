package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"

	"Stage4/internal/handlers"
	"Stage4/internal/repository"
)

func main() {
	// Открываем базу (файл создастся автоматически)
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS expenses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	amount REAL NOT NULL,
	description TEXT NOT NULL,
	date TEXT NOT NULL)`)
	if err != nil {
		log.Fatal("миграция:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE)`)
	if err != nil {
		log.Fatal("миграция categories:", err)
	}

	// Базовые категории (один раз; INSERT OR IGNORE не создаст дублей)
	_, err = db.Exec(`INSERT OR IGNORE INTO categories (name) VALUES
	('Еда'), ('Транспорт'), ('Жильё'), ('Развлечения'), ('Прочее')`)
	if err != nil {
		log.Fatal("seed categories:", err)
	}

	// Поле category_id в expenses
	_, err = db.Exec(`ALTER TABLE expenses ADD COLUMN category_id INTEGER REFERENCES categories(id)`)
	if err != nil {
		// ALTER TABLE падает, если колонка уже есть — это нормально при повторном запуске.
		// В учебном проекте можно просто залогировать и продолжить.
		log.Println("ALTER expenses:", err)
	}

	repo := repository.NewExpenseRepository(db)
	catRepo := repository.NewCategoryRepository(db)
	h := &handlers.ExpenseHandler{
		Repo:    repo,
		CatRepo: catRepo,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/expenses", h.List)
	mux.HandleFunc("/expenses/add", h.Add)
	mux.HandleFunc("/expenses/edit", h.EditForm)
	mux.HandleFunc("/expenses/update", h.Update)
	mux.HandleFunc("/expenses/delete", h.Delete)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
