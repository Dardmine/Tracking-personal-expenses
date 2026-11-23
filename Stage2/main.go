package main

import (
	"log"
	"net/http"

	"Stage2/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/expenses", handlers.ListExpenses)
	mux.HandleFunc("/expenses/add", handlers.AddExpense)

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}