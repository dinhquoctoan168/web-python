package database

import (
	"database/sql"
	"encoding/json"
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
	{
		Version: 4,
		Name:    "classes",
		Up: func(tx *sql.Tx) error {
			queries := []string{
				`CREATE TABLE IF NOT EXISTS classes (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					course_id INTEGER NOT NULL,
					name TEXT NOT NULL,
					semester TEXT,
					academic_year TEXT,
					teacher_id INTEGER NOT NULL,
					status TEXT DEFAULT 'active',
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY(course_id) REFERENCES courses(id),
					FOREIGN KEY(teacher_id) REFERENCES users(id)
				);`,
				`CREATE TABLE IF NOT EXISTS enrollments (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					class_id INTEGER NOT NULL,
					student_id INTEGER NOT NULL,
					enrolled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					UNIQUE(class_id, student_id),
					FOREIGN KEY(class_id) REFERENCES classes(id),
					FOREIGN KEY(student_id) REFERENCES users(id)
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
		Version: 5,
		Name:    "lessons",
		Up: func(tx *sql.Tx) error {
			queries := []string{
				`CREATE TABLE IF NOT EXISTS chapters (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					course_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					description TEXT,
					order_num INTEGER DEFAULT 0,
					FOREIGN KEY(course_id) REFERENCES courses(id)
				);`,
				`CREATE TABLE IF NOT EXISTS lessons (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					chapter_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					content_html TEXT,
					order_num INTEGER DEFAULT 0,
					is_published INTEGER DEFAULT 0,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY(chapter_id) REFERENCES chapters(id)
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
		Version: 6,
		Name:    "exercises",
		Up: func(tx *sql.Tx) error {
			// 1. Đổi tên exercises cũ thành exercises_old
			if _, err := tx.Exec(`ALTER TABLE exercises RENAME TO exercises_old;`); err != nil {
				return err
			}

			// 2. Tạo bảng exercises mới
			createExercises := `CREATE TABLE exercises (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				course_id INTEGER NOT NULL DEFAULT 3,
				lesson_id INTEGER,
				topic_id INTEGER,
				title TEXT NOT NULL,
				exercise_type TEXT NOT NULL DEFAULT 'coding',
				difficulty TEXT DEFAULT 'Dễ',
				description TEXT NOT NULL,
				initial_code TEXT,
				solution_code TEXT,
				solution_hint TEXT,
				allowed_functions TEXT,
				time_limit_ms INTEGER DEFAULT 5000,
				status TEXT DEFAULT 'active',
				created_by INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY(course_id) REFERENCES courses(id),
				FOREIGN KEY(lesson_id) REFERENCES lessons(id)
			);`
			if _, err := tx.Exec(createExercises); err != nil {
				return err
			}

			// 3. Tạo bảng exercise_test_cases
			createTestCases := `CREATE TABLE exercise_test_cases (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				exercise_id INTEGER NOT NULL,
				input_data TEXT,
				call_expression TEXT,
				expected_output TEXT,
				is_hidden INTEGER DEFAULT 0,
				weight REAL DEFAULT 1,
				order_num INTEGER DEFAULT 0,
				FOREIGN KEY(exercise_id) REFERENCES exercises(id)
			);`
			if _, err := tx.Exec(createTestCases); err != nil {
				return err
			}

			// 4. Sao chép dữ liệu bài tập từ exercises_old sang exercises mới
			copyExercises := `INSERT INTO exercises (id, course_id, topic_id, title, difficulty, description, initial_code, allowed_functions, solution_hint, created_at)
			SELECT id, 3, topic_id, title, difficulty, description, initial_code, allowed_functions, solution_hint, created_at
			FROM exercises_old;`
			if _, err := tx.Exec(copyExercises); err != nil {
				return err
			}

			// 5. Chuyển đổi test_cases JSON sang bảng quan hệ exercise_test_cases
			rows, err := tx.Query(`SELECT id, test_cases FROM exercises_old`)
			if err != nil {
				return err
			}
			defer rows.Close()

			type oldTestCase struct {
				Call     string `json:"call"`
				Input    string `json:"input"`
				Expected string `json:"expected"`
			}

			type exRow struct {
				id  int
				raw string
			}
			var exRows []exRow
			for rows.Next() {
				var item exRow
				if err := rows.Scan(&item.id, &item.raw); err != nil {
					return err
				}
				exRows = append(exRows, item)
			}
			rows.Close()

			insertTC := `INSERT INTO exercise_test_cases (exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) VALUES (?, ?, ?, ?, ?, ?, ?)`
			for _, item := range exRows {
				var tcs []oldTestCase
				if err := json.Unmarshal([]byte(item.raw), &tcs); err == nil {
					for idx, tc := range tcs {
						if _, err := tx.Exec(insertTC, item.id, tc.Input, tc.Call, tc.Expected, 0, 1.0, idx+1); err != nil {
							return err
						}
					}
				}

				// Bổ sung hidden test case tương ứng cho từng bài
				switch item.id {
				case 1:
					_, _ = tx.Exec(insertTC, item.id, "[100, 2, 50, -99, 0, 1000]", "find_min_max([100, 2, 50, -99, 0, 1000])", "(-99, 1000)", 1, 1.0, 99)
				case 2:
					_, _ = tx.Exec(insertTC, item.id, `"{[()]}"`, `is_valid_parentheses("{[()]}")`, "True", 1, 1.0, 99)
				case 3:
					_, _ = tx.Exec(insertTC, item.id, "students, min_gpa=4.0, dept='CNTT'", "query_students(students, 4.0, 'CNTT')", "[]", 1, 1.0, 99)
				case 4:
					_, _ = tx.Exec(insertTC, item.id, "arr=[1, 3, 5, 7, 9], target=1", "binary_search([1, 3, 5, 7, 9], 1)", "0", 1, 1.0, 99)
				}
			}

			// 6. Xóa bảng exercises_old
			if _, err := tx.Exec(`DROP TABLE exercises_old;`); err != nil {
				return err
			}

			return nil
		},
	},
	{
		Version: 7,
		Name:    "practice",
		Up: func(tx *sql.Tx) error {
			query := `CREATE TABLE IF NOT EXISTS student_exercise_progress (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				student_id INTEGER NOT NULL,
				exercise_id INTEGER NOT NULL,
				status TEXT NOT NULL DEFAULT 'not_started',
				last_code TEXT,
				best_score REAL DEFAULT 0,
				attempts INTEGER DEFAULT 0,
				first_started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				completed_at DATETIME,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY(student_id) REFERENCES users(id),
				FOREIGN KEY(exercise_id) REFERENCES exercises(id),
				UNIQUE(student_id, exercise_id)
			);`
			_, err := tx.Exec(query)
			return err
		},
	},
	{
		Version: 8,
		Name:    "submissions",
		Up: func(tx *sql.Tx) error {
			queries := []string{
				`CREATE TABLE IF NOT EXISTS submissions (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					student_id INTEGER NOT NULL,
					exercise_id INTEGER NOT NULL,
					source_code TEXT NOT NULL,
					score REAL DEFAULT 0,
					passed_tests INTEGER DEFAULT 0,
					total_tests INTEGER DEFAULT 0,
					status TEXT,
					submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY(student_id) REFERENCES users(id),
					FOREIGN KEY(exercise_id) REFERENCES exercises(id)
				);`,
				`CREATE TABLE IF NOT EXISTS exercise_attempts (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					student_id INTEGER NOT NULL,
					exercise_id INTEGER NOT NULL,
					started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					ended_at DATETIME,
					run_count INTEGER DEFAULT 0,
					test_count INTEGER DEFAULT 0,
					hint_count INTEGER DEFAULT 0,
					FOREIGN KEY(student_id) REFERENCES users(id),
					FOREIGN KEY(exercise_id) REFERENCES exercises(id)
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
		Version: 9,
		Name:    "assignments",
		Up: func(tx *sql.Tx) error {
			queries := []string{
				`CREATE TABLE IF NOT EXISTS assignments (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					class_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					description TEXT,
					start_at DATETIME,
					due_at DATETIME,
					status TEXT DEFAULT 'draft',
					created_by INTEGER,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY(class_id) REFERENCES classes(id),
					FOREIGN KEY(created_by) REFERENCES users(id)
				);`,
				`CREATE TABLE IF NOT EXISTS assignment_exercises (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					assignment_id INTEGER NOT NULL,
					exercise_id INTEGER NOT NULL,
					points REAL DEFAULT 1,
					order_num INTEGER DEFAULT 0,
					FOREIGN KEY(assignment_id) REFERENCES assignments(id),
					FOREIGN KEY(exercise_id) REFERENCES exercises(id)
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
