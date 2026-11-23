package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"Stage2/internal/models"
)

var expenses []models.Expense
var nextID = 1

// Парсим шаблоны один раз при старте
var tmpl = template.Must(template.ParseGlob("web/templates/*.html"))

func ListExpenses(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "layout", expenses); err != nil {
		log.Println("template error:", err)
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func ShowAddForm(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "expenses.html", expenses)
}

// POST-форма!
func AddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}

	expenses = append(expenses, models.Expense{
		ID:          nextID,
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	})
	nextID++

	// Post/Redirect/Get: после POST делаем redirect на список
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}
