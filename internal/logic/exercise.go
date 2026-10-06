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

// FunctionItem đại diện cho thông tin 1 hàm tra cứu
type FunctionItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Syntax      string `json:"syntax"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// GetTopicsWithExercises lấy toàn bộ chủ đề kèm danh sách bài tập tương ứng
func GetTopicsWithExercises() ([]Topic, error) {
	db := database.GetDB()
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
			// Parse JSON danh sách hàm cho phép
			_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)

			// Lấy public test cases (is_hidden = 0)
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
