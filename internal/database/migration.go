package database

import (
	"database/sql"
	"fmt"
	"log"
)

// Migration định nghĩa một bước di chuyển schema
type Migration struct {
	Version int
	Name    string
	Up      func(tx *sql.Tx) error
}

// migrations chứa danh sách các migration theo thứ tự phiên bản
var migrations = []Migration{
	{
		Version: 1,
		Name:    "initial",
		Up: func(tx *sql.Tx) error {
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
				if _, err := tx.Exec(q); err != nil {
					return err
				}
			}
			return nil
		},
	},
	{
		Version: 2,
		Name:    "users",
		Up: func(tx *sql.Tx) error {
			queries := []string{
				`CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					username TEXT NOT NULL UNIQUE,
					password_hash TEXT NOT NULL,
					full_name TEXT NOT NULL,
					email TEXT,
					role TEXT NOT NULL,
					is_active INTEGER DEFAULT 1,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);`,
				`CREATE TABLE IF NOT EXISTS sessions (
					id TEXT PRIMARY KEY,
					user_id INTEGER NOT NULL,
					expires_at DATETIME NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY(user_id) REFERENCES users(id)
				);`,
			}
			for _, q := range queries {
				if _, err := tx.Exec(q); err != nil {
					return err
				}
			}
			return nil
		},
	},
	{
		Version: 3,
		Name:    "courses",
		Up: func(tx *sql.Tx) error {
			query := `CREATE TABLE IF NOT EXISTS courses (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				code TEXT NOT NULL UNIQUE,
				name TEXT NOT NULL,
				description TEXT,
				status TEXT DEFAULT 'active',
				created_by INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY(created_by) REFERENCES users(id)
			);`
			_, err := tx.Exec(query)
			return err
		},
	},
}

// RunMigrations thực thi các migration chưa được áp dụng
func RunMigrations(db *sql.DB) error {
	// 1. Đảm bảo bảng schema_migrations tồn tại
	createTableQuery := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("không thể tạo bảng schema_migrations: %w", err)
	}

	// 2. Lấy danh sách các phiên bản đã áp dụng
	applied := make(map[int]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("không thể đọc schema_migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// 3. Thực thi từng migration chưa áp dụng trong transaction
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("không thể khởi tạo transaction cho migration %d: %w", m.Version, err)
		}

		if err := m.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d (%s) thất bại: %w", m.Version, m.Name, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.Version, m.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("lỗi ghi nhận migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("lỗi commit migration %d: %w", m.Version, err)
		}

		log.Printf("Đã áp dụng migration: %03d_%s", m.Version, m.Name)
	}

	return nil
}
