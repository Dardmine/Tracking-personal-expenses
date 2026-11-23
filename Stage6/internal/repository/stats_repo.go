package repository

import (
	"database/sql"
	"log"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

type CategoryStat struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type MonthStat struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

func (r *StatsRepository) ByCategory(userID int, from, to string) ([]CategoryStat, error) {
	query := `SELECT COALESCE(c.name, 'Без категории') AS name, SUM(e.amount) AS total
	          FROM expenses e
	          LEFT JOIN categories c ON c.id = e.category_id
	          WHERE e.user_id = ?`
	args := []interface{}{userID}

	if from != "" {
		query += " AND e.date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND e.date <= ?"
		args = append(args, to)
	}
	query += " GROUP BY c.name ORDER BY total DESC"

	log.Printf("ByCategory: userID=%d, from=%q, to=%q", userID, from, to)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CategoryStat
	for rows.Next() {
		var s CategoryStat
		if err := rows.Scan(&s.Name, &s.Amount); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *StatsRepository) ByMonth(userID int, year string) ([]MonthStat, error) {
	rows, err := r.db.Query(`
		SELECT strftime('%Y-%m', date) AS month, SUM(amount) AS total
		FROM expenses
		WHERE user_id = ? AND strftime('%Y', date) = ?
		GROUP BY month
		ORDER BY month`, userID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MonthStat
	for rows.Next() {
		var s MonthStat
		if err := rows.Scan(&s.Month, &s.Amount); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *StatsRepository) AvailableYears(userID int) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT strftime('%Y', date) AS year
		FROM expenses
		WHERE user_id = ? AND date IS NOT NULL AND date <> ''
		ORDER BY year DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []string
	for rows.Next() {
		var y sql.NullString
		if err := rows.Scan(&y); err != nil {
			return nil, err
		}
		if y.Valid && y.String != "" {
			years = append(years, y.String)
		}
	}
	return years, nil
}
