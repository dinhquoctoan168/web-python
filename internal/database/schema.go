package database

import "database/sql"

func initSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS topics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			order_num INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			topic_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			difficulty TEXT DEFAULT 'Dễ',
			description TEXT NOT NULL,
			initial_code TEXT NOT NULL,
			allowed_functions TEXT NOT NULL,
			test_cases TEXT NOT NULL,
			solution_hint TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (topic_id) REFERENCES topics(id)
		);`,
		`CREATE TABLE IF NOT EXISTS functions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			category TEXT NOT NULL,
			syntax TEXT NOT NULL,
			description TEXT NOT NULL,
			example TEXT NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
