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
	AllowedFunctions []string `json:"allowed_functions"`
	TestCasesJSON    string   `json:"test_cases_json"`
	SolutionHint     string   `json:"solution_hint"`
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
		exRows, err := db.Query(`SELECT id, topic_id, title, difficulty, description, initial_code, allowed_functions, test_cases, solution_hint 
			FROM exercises WHERE topic_id = ? ORDER BY id ASC`, topics[i].ID)
		if err != nil {
			return nil, err
		}

		var exercises []Exercise
		for exRows.Next() {
			var ex Exercise
			var allowedFuncsRaw string
			if err := exRows.Scan(&ex.ID, &ex.TopicID, &ex.Title, &ex.Difficulty, &ex.Description, &ex.InitialCode, &allowedFuncsRaw, &ex.TestCasesJSON, &ex.SolutionHint); err != nil {
				exRows.Close()
				return nil, err
			}
			// Parse JSON danh sách hàm cho phép
			_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)
			exercises = append(exercises, ex)
		}
		exRows.Close()
		topics[i].Exercises = exercises
	}

	return topics, nil
}

// GetExerciseByID lấy chi tiết 1 bài tập theo id
func GetExerciseByID(id int) (*Exercise, error) {
	db := database.GetDB()
	var ex Exercise
	var allowedFuncsRaw string

	query := `SELECT e.id, e.topic_id, t.name, e.title, e.difficulty, e.description, e.initial_code, e.allowed_functions, e.test_cases, e.solution_hint 
		FROM exercises e 
		JOIN topics t ON e.topic_id = t.id 
		WHERE e.id = ?`

	err := db.QueryRow(query, id).Scan(
		&ex.ID, &ex.TopicID, &ex.TopicName, &ex.Title, &ex.Difficulty,
		&ex.Description, &ex.InitialCode, &allowedFuncsRaw, &ex.TestCasesJSON, &ex.SolutionHint,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)
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
