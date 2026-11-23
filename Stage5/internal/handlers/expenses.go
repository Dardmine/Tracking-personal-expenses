package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"Stage5/internal/middleware"
	"Stage5/internal/models"
	"Stage5/internal/repository"
)

type ExpenseHandler struct {
	Repo    *repository.ExpenseRepository
	CatRepo *repository.CategoryRepository
}

var tmpl = template.Must(template.ParseGlob("web/templates/*.html"))

// List — список трат текущего пользователя с фильтрами.
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	q := r.URL.Query()
	catID, _ := strconv.Atoi(q.Get("category_id"))
	from := q.Get("from")
	to := q.Get("to")

	items, err := h.Repo.GetFiltered(userID, catID, from, to) // ← userID первым
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

// Add — добавление траты текущему пользователю.
func (h *ExpenseHandler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int)

	r.ParseForm()
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}
	catID, _ := strconv.Atoi(r.FormValue("category_id"))

	err = h.Repo.Create(models.Expense{
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
		CategoryID:  catID,
	}, userID) // ← userID вторым аргументом
	if err != nil {
		log.Println("create error:", err)
		http.Error(w, "Ошибка записи", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// EditForm — форма редактирования своей траты.
func (h *ExpenseHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	e, err := h.Repo.GetByID(id, userID) // ← userID вторым аргументом
	if err != nil {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout_edit", e); err != nil {
		log.Println("template error:", err)
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Update — сохранение изменений своей траты.
func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	catID, _ := strconv.Atoi(r.FormValue("category_id"))

	err := h.Repo.Update(models.Expense{
		ID:          id,
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
		CategoryID:  catID,
	}, userID) // ← userID вторым аргументом
	if err != nil {
		log.Println("update error:", err)
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// Delete — удаление своей траты.
func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if err := h.Repo.Delete(id, userID); err != nil { // ← userID вторым аргументом
		log.Println("delete error:", err)
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}
