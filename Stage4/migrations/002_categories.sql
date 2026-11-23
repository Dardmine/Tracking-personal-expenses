CREATE TABLE IF NOT EXISTS categories (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

INSERT OR IGNORE INTO categories (name) VALUES
    ('Еда'), ('Транспорт'), ('Жильё'), ('Развлечения'), ('Прочее');

-- Добавляем поле category_id в expenses (nullable)
ALTER TABLE expenses ADD COLUMN category_id INTEGER REFERENCES categories(id);