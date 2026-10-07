// Package logic
// Mục đích: Cung cấp logic nghiệp vụ truy vấn bài tập và danh mục hàm môn học.
package logic

import (
	"database/sql"
	"encoding/json"
	"strings"

	"web_python/internal/database"
)

// Exercise đại diện cho 1 bài tập
type Exercise struct {
	ID               int      `json:"id"`
	TopicID          int      `json:"topic_id"`
	TopicName        string   `json:"topic_name,omitempty"`
	Title            string   `json:"title"`
	Difficulty       string   `json:"difficulty"`
	Description      string   `json:"description"`
	InitialCode      string   `json:"initial_code"`
	AllowedFunctions  []string `json:"allowed_functions"`
	TestCasesJSON     string   `json:"test_cases_json"`
	SolutionHint      string   `json:"solution_hint"`
	VisualizationType string   `json:"visualization_type,omitempty"`
}

// Topic đại diện cho 1 chủ đề
type Topic struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	OrderNum    int        `json:"order_num"`
	Exercises   []Exercise `json:"exercises"`
}

// ChapterItem đại diện cho một chương học kèm các bài tập thuộc chương đó (Phase 24)
type ChapterItem struct {
	ID          int        `json:"id"`
	CourseID    int        `json:"course_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	OrderNum    int        `json:"order_num"`
	Exercises   []Exercise `json:"exercises"`
}

// FunctionItem đại diện cho thông tin 1 hàm tra cứu
type FunctionItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Syntax      string `json:"syntax"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// GetCourseStructure lấy cấu trúc môn học gồm các chương và bài tập (thay thế GetTopicsWithExercises theo Phase 24)
func GetCourseStructure(courseCode string) ([]ChapterItem, error) {
	db := database.GetDB()
	if db == nil {
		return nil, nil
	}

	courseCode = strings.TrimSpace(courseCode)
	if courseCode == "" {
		courseCode = "DSA301"
	}

	var courseID int
	err := db.QueryRow(`SELECT id FROM courses WHERE code = ? LIMIT 1`, courseCode).Scan(&courseID)
	if err != nil {
		_ = db.QueryRow(`SELECT id FROM courses ORDER BY id ASC LIMIT 1`).Scan(&courseID)
	}

	if courseID == 0 {
		return nil, nil
	}

	cRows, err := db.Query(`SELECT id, course_id, title, COALESCE(description, ''), order_num 
		FROM chapters WHERE course_id = ? ORDER BY order_num ASC, id ASC`, courseID)
	if err != nil {
		return nil, err
	}
	defer cRows.Close()

	var chapters []ChapterItem
	for cRows.Next() {
		var c ChapterItem
		if err := cRows.Scan(&c.ID, &c.CourseID, &c.Title, &c.Description, &c.OrderNum); err == nil {
			chapters = append(chapters, c)
		}
	}

	for i := range chapters {
		chapID := chapters[i].ID
		exRows, err := db.Query(`SELECT e.id, COALESCE(e.topic_id, 0), e.title, e.difficulty, e.description, 
			e.initial_code, COALESCE(e.allowed_functions, '[]'), COALESCE(e.solution_hint, ''), COALESCE(e.visualization_type, '')
			FROM exercises e
			LEFT JOIN lessons l ON e.lesson_id = l.id
			WHERE l.chapter_id = ? OR (e.course_id = ? AND e.lesson_id IS NULL)
			ORDER BY e.id ASC`, chapID, courseID)
		if err != nil {
			continue
		}

		var exList []Exercise
		for exRows.Next() {
			var ex Exercise
			var allowedFuncsRaw string
			if err := exRows.Scan(&ex.ID, &ex.TopicID, &ex.Title, &ex.Difficulty, &ex.Description, &ex.InitialCode, &allowedFuncsRaw, &ex.SolutionHint, &ex.VisualizationType); err == nil {
				_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)

				// Lấy public test cases
				tcRows, tcErr := db.Query(`SELECT input_data, call_expression, expected_output 
					FROM exercise_test_cases WHERE exercise_id = ? AND is_hidden = 0 ORDER BY order_num ASC, id ASC`, ex.ID)
				if tcErr == nil {
					type clientTC struct {
						Call     string `json:"call"`
						Input    string `json:"input"`
						Expected string `json:"expected"`
					}
					var tcs []clientTC
					for tcRows.Next() {
						var inp, call, exp string
						if err := tcRows.Scan(&inp, &call, &exp); err == nil {
							tcs = append(tcs, clientTC{Call: call, Input: inp, Expected: exp})
						}
					}
					tcRows.Close()
					tcBytes, _ := json.Marshal(tcs)
					ex.TestCasesJSON = string(tcBytes)
				} else {
					ex.TestCasesJSON = "[]"
				}

				exList = append(exList, ex)
			}
		}
		exRows.Close()
		chapters[i].Exercises = exList
	}

	return chapters, nil
}

// GetTopicsWithExercises lấy toàn bộ chủ đề kèm danh sách bài tập tương ứng
// Deprecated: Đã thay thế bằng GetCourseStructure theo Phase 24. Hàm này được giữ lại để tương thích ngược.
func GetTopicsWithExercises() ([]Topic, error) {
	db := database.GetDB()
	if db == nil {
		return nil, nil
	}

	var topicCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM topics").Scan(&topicCount)
	if topicCount == 0 {
		// Tự động chuyển tiếp sang cấu trúc chapters mới nếu bảng topics không có dữ liệu
		chaps, err := GetCourseStructure("DSA301")
		if err != nil {
			return nil, err
		}
		var topics []Topic
		for _, c := range chaps {
			topics = append(topics, Topic{
				ID:          c.ID,
				Name:        c.Title,
				Description: c.Description,
				OrderNum:    c.OrderNum,
				Exercises:   c.Exercises,
			})
		}
		return topics, nil
	}

	rows, err := db.Query("SELECT id, name, description, icon, order_num FROM topics ORDER BY order_num ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var t Topic
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Icon, &t.OrderNum); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}

	for i := range topics {
		exRows, err := db.Query(`SELECT id, topic_id, title, difficulty, description, initial_code, allowed_functions, COALESCE(solution_hint, ''), COALESCE(visualization_type, '') 
			FROM exercises WHERE topic_id = ? ORDER BY id ASC`, topics[i].ID)
		if err != nil {
			return nil, err
		}

		var exercises []Exercise
		for exRows.Next() {
			var ex Exercise
			var allowedFuncsRaw string
			if err := exRows.Scan(&ex.ID, &ex.TopicID, &ex.Title, &ex.Difficulty, &ex.Description, &ex.InitialCode, &allowedFuncsRaw, &ex.SolutionHint, &ex.VisualizationType); err != nil {
				exRows.Close()
				return nil, err
			}
			_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)

			tcRows, tcErr := db.Query(`SELECT input_data, call_expression, expected_output 
				FROM exercise_test_cases WHERE exercise_id = ? AND is_hidden = 0 ORDER BY order_num ASC, id ASC`, ex.ID)
			if tcErr == nil {
				type clientTC struct {
					Call     string `json:"call"`
					Input    string `json:"input"`
					Expected string `json:"expected"`
				}
				var tcs []clientTC
				for tcRows.Next() {
					var inp, call, exp string
					if err := tcRows.Scan(&inp, &call, &exp); err == nil {
						tcs = append(tcs, clientTC{Call: call, Input: inp, Expected: exp})
					}
				}
				tcRows.Close()
				tcBytes, _ := json.Marshal(tcs)
				ex.TestCasesJSON = string(tcBytes)
			} else {
				ex.TestCasesJSON = "[]"
			}

			exercises = append(exercises, ex)
		}
		exRows.Close()
		topics[i].Exercises = exercises
	}

	return topics, nil
}

// GetExerciseByID lấy chi tiết 1 bài tập theo id (CHỈ trả public test cases, KHÔNG trả hidden tests hay solution code)
func GetExerciseByID(id int) (*Exercise, error) {
	db := database.GetDB()
	var ex Exercise
	var allowedFuncsRaw string

	query := `SELECT e.id, e.topic_id, COALESCE(t.name, ''), e.title, e.difficulty, e.description, e.initial_code, 
		COALESCE(e.allowed_functions, '[]'), COALESCE(e.solution_hint, ''), COALESCE(e.visualization_type, '') 
		FROM exercises e 
		LEFT JOIN topics t ON e.topic_id = t.id 
		WHERE e.id = ?`

	err := db.QueryRow(query, id).Scan(
		&ex.ID, &ex.TopicID, &ex.TopicName, &ex.Title, &ex.Difficulty,
		&ex.Description, &ex.InitialCode, &allowedFuncsRaw, &ex.SolutionHint, &ex.VisualizationType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)

	// Truy vấn các public test cases (is_hidden = 0)
	tcRows, err := db.Query(`SELECT input_data, call_expression, expected_output 
		FROM exercise_test_cases WHERE exercise_id = ? AND is_hidden = 0 ORDER BY order_num ASC, id ASC`, ex.ID)
	if err == nil {
		type clientTC struct {
			Call     string `json:"call"`
			Input    string `json:"input"`
			Expected string `json:"expected"`
		}
		var tcs []clientTC
		for tcRows.Next() {
			var inp, call, exp string
			if err := tcRows.Scan(&inp, &call, &exp); err == nil {
				tcs = append(tcs, clientTC{Call: call, Input: inp, Expected: exp})
			}
		}
		tcRows.Close()
		tcBytes, _ := json.Marshal(tcs)
		ex.TestCasesJSON = string(tcBytes)
	} else {
		ex.TestCasesJSON = "[]"
	}

	return &ex, nil
}

// SearchFunctions tìm kiếm hàm theo từ khoá hoặc lấy tất cả nếu query rỗng
func SearchFunctions(query string) ([]FunctionItem, error) {
	db := database.GetDB()
	var rows *sql.Rows
	var err error

	query = strings.TrimSpace(query)
	if query == "" {
		rows, err = db.Query("SELECT id, name, category, syntax, description, example FROM functions ORDER BY category, name ASC")
	} else {
		like := "%" + query + "%"
		rows, err = db.Query("SELECT id, name, category, syntax, description, example FROM functions WHERE name LIKE ? OR category LIKE ? OR description LIKE ? ORDER BY name ASC", like, like, like)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FunctionItem
	for rows.Next() {
		var item FunctionItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Syntax, &item.Description, &item.Example); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
