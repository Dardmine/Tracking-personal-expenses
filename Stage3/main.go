package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"

	"Stage3/internal/handlers"
	"Stage3/internal/repository"
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

	repo := repository.NewExpenseRepository(db)
	h := &handlers.ExpenseHandler{Repo: repo}

	mux := http.NewServeMux()
	mux.HandleFunc("/expenses", h.List)
	mux.HandleFunc("/expenses/add", h.Add)
	mux.HandleFunc("/expenses/edit", h.EditForm)
	mux.HandleFunc("/expenses/update", h.Update)
	mux.HandleFunc("/expenses/delete", h.Delete)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
