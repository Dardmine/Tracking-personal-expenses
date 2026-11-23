package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"Stage6/internal/middleware"
	"Stage6/internal/repository"
)

type StatsHandler struct {
	Repo *repository.StatsRepository
}

var statsTmpl = template.Must(template.ParseGlob("web/templates/*.html"))

// HTML-страница статистики с выбором года и периода.
func (h *StatsHandler) Page(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	years, err := h.Repo.AvailableYears(userID)
	if err != nil {
		log.Println("years error:", err)
		http.Error(w, "Ошибка чтения БД", http.StatusInternalServerError)
		return
	}

	selectedYear := r.URL.Query().Get("year")
	if selectedYear == "" {
		if len(years) > 0 {
			selectedYear = years[0]
		}
	}

	data := map[string]interface{}{
		"Years":        years,
		"SelectedYear": selectedYear,
	}
	if err := statsTmpl.ExecuteTemplate(w, "layout_stats", data); err != nil {
		log.Println("template error:", err)
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *StatsHandler) ByCategory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")

	data, err := h.Repo.ByCategory(userID, from, to)
	if err != nil {
		log.Println("stats by-category error:", err)
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *StatsHandler) ByMonth(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	year := r.URL.Query().Get("year")
	if year == "" {
		http.Error(w, "Не указан год", http.StatusBadRequest)
		return
	}

	data, err := h.Repo.ByMonth(userID, year)
	if err != nil {
		log.Println("stats by-month error:", err)
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
