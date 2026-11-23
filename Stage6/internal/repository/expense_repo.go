package repository

import (
	"database/sql"

	"Stage6/internal/models"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(e models.Expense, userID int) error {
	_, err := r.db.Exec(
		`INSERT INTO expenses (amount, description, date, category_id, user_id)
		 VALUES (?, ?, ?, ?, ?)`,
		e.Amount, e.Description, e.Date, e.CategoryID, userID,
	)
	return err
}

func (r *ExpenseRepository) GetAll(userID int) ([]models.Expense, error) {
	return r.GetFiltered(userID, 0, "", "")
}

func (r *ExpenseRepository) GetFiltered(userID, categoryID int, from, to string) ([]models.Expense, error) {
	query := `SELECT e.id, e.amount, e.description, e.date,
	                 COALESCE(e.category_id, 0),
	                 COALESCE(c.name, 'Без категории')
	          FROM expenses e
	          LEFT JOIN categories c ON c.id = e.category_id
	          WHERE e.user_id = ?`
	args := []interface{}{userID}

	if categoryID > 0 {
		query += " AND e.category_id = ?"
		args = append(args, categoryID)
	}
	if from != "" {
		query += " AND e.date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND e.date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY e.date DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Date,
			&e.CategoryID, &e.CategoryName); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *ExpenseRepository) GetByID(id, userID int) (models.Expense, error) {
	var e models.Expense
	err := r.db.QueryRow(
		`SELECT e.id, e.amount, e.description, e.date,
		        COALESCE(e.category_id, 0), COALESCE(c.name, 'Без категории')
		 FROM expenses e
		 LEFT JOIN categories c ON c.id = e.category_id
		 WHERE e.id = ? AND e.user_id = ?`,
		id, userID,
	).Scan(&e.ID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.CategoryName)
	return e, err
}

func (r *ExpenseRepository) Update(e models.Expense, userID int) error {
	_, err := r.db.Exec(
		`UPDATE expenses
		 SET amount = ?, description = ?, date = ?, category_id = ?
		 WHERE id = ? AND user_id = ?`,
		e.Amount, e.Description, e.Date, e.CategoryID, e.ID, userID,
	)
	return err
}

func (r *ExpenseRepository) Delete(id, userID int) error {
	_, err := r.db.Exec(
		"DELETE FROM expenses WHERE id = ? AND user_id = ?",
		id, userID,
	)
	return err
}
