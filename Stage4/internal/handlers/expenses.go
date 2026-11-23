package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"Stage4/internal/models"
	"Stage4/internal/repository"
)

// ExpenseHandler — структура с зависимостью от репозитория
type ExpenseHandler struct {
	Repo    *repository.ExpenseRepository
	CatRepo *repository.CategoryRepository
}

// Все шаблоны парсим один раз при старте
var tmpl = template.Must(template.ParseGlob("web/templates/*.html"))

// func Templates() []*template.Template { return tmpl.Templates() }

// List — список трат (использует layout_list)
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	catID, _ := strconv.Atoi(q.Get("category_id"))
	from := q.Get("from")
	to := q.Get("to")

	items, err := h.Repo.GetFiltered(catID, from, to)
	if err != nil {
		log.Println("repo error:", err)
		http.Error(w, "Ошибка чтения БД", http.StatusInternalServerError)
		return
	}
	cats, err := h.CatRepo.GetAll()
	if err != nil {
		log.Println("cat repo error:", err)
		http.Error(w, "Ошибка категорий", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Items":      items,
		"Categories": cats,
		"Selected":   catID,
		"From":       from,
		"To":         to,
	}
	if err := tmpl.ExecuteTemplate(w, "layout_list", data); err != nil {
		log.Println("template error:", err)
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Add — обработка POST-формы добавления
func (h *ExpenseHandler) Add(w http.ResponseWriter, r *http.Request) {
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
	catID, _ := strconv.Atoi(r.FormValue("category_id"))
	log.Printf("Add: amount=%v desc=%q date=%q catID=%d",
		amount, r.FormValue("description"), r.FormValue("date"), catID)

	err = h.Repo.Create(models.Expense{
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
		CategoryID:  catID,
	})
	if err != nil {
		log.Println("create error:", err)
		http.Error(w, "Ошибка записи", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// EditForm — страница редактирования (использует layout_edit)
func (h *ExpenseHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	e, err := h.Repo.GetByID(id)
	if err != nil {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout_edit", e); err != nil {
		log.Println("template error:", err)
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Update — обработка POST-формы редактирования
func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	err := h.Repo.Update(models.Expense{
		ID:          id,
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	})
	if err != nil {
		log.Println("update error:", err)
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// Delete — удаление траты
func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if err := h.Repo.Delete(id); err != nil {
		log.Println("delete error:", err)
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}
