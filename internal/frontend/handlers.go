// Package frontend
// Mục đích: Định nghĩa các HTTP handler và render giao diện HTML qua thư viện chuẩn html/template.
package frontend

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"web_python/internal/logic"
	"web_python/internal/security"
)

// IDEPageData chứa dữ liệu truyền vào template HTML của IDE
type IDEPageData struct {
	Title           string
	CSRFToken       string
	CourseCode      string
	Chapters        []logic.ChapterItem
	Topics          []logic.Topic // Tương thích ngược với các template cũ
	CurrentExercise *logic.Exercise
	Functions       []logic.FunctionItem
}

// HandleIDE render trang giao diện lập trình IDE
func HandleIDE(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/ide" {
		http.NotFound(w, r)
		return
	}

	courseCode := r.URL.Query().Get("course")
	if courseCode == "" {
		courseCode = "DSA301"
	}

	chapters, err := logic.GetCourseStructure(courseCode)
	if err != nil {
		log.Printf("Lỗi lấy cấu trúc môn học %s: %v", courseCode, err)
	}

	// Đọc danh sách topics cũ hoặc adapter
	topics, _ := logic.GetTopicsWithExercises()

	// Xác định bài tập hiện tại
	var currentExercise *logic.Exercise
	exIDStr := r.URL.Query().Get("id")
	if exIDStr != "" {
		if id, err := strconv.Atoi(exIDStr); err == nil {
			currentExercise, _ = logic.GetExerciseByID(id)
		}
	}

	// Mặc định chọn bài đầu tiên nếu chưa chọn
	if currentExercise == nil {
		if len(chapters) > 0 && len(chapters[0].Exercises) > 0 {
			firstID := chapters[0].Exercises[0].ID
			currentExercise, _ = logic.GetExerciseByID(firstID)
		} else if len(topics) > 0 && len(topics[0].Exercises) > 0 {
			firstID := topics[0].Exercises[0].ID
			currentExercise, _ = logic.GetExerciseByID(firstID)
		}
	}

	funcs, err := logic.SearchFunctions("")
	if err != nil {
		log.Printf("Lỗi lấy danh sách functions: %v", err)
	}

	data := IDEPageData{
		Title:           "Web Python IDE - CSDL & Giải thuật",
		CSRFToken:       security.GetTokenFromContext(r.Context()),
		CourseCode:      courseCode,
		Chapters:        chapters,
		Topics:          topics,
		CurrentExercise: currentExercise,
		Functions:       funcs,
	}

	tmpl, err := template.ParseFiles("web/templates/base.html", "web/templates/ide.html")
	if err != nil {
		log.Printf("Lỗi parse template: %v", err)
		http.Error(w, "Lỗi hiển thị giao diện", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Lỗi execute template: %v", err)
	}
}

// HandleAPIExercise trả về chi tiết bài tập theo dạng JSON
func HandleAPIExercise(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID bài tập không hợp lệ", http.StatusBadRequest)
		return
	}

	ex, err := logic.GetExerciseByID(id)
	if err != nil {
		http.Error(w, "Lỗi truy vấn bài tập", http.StatusInternalServerError)
		return
	}
	if ex == nil {
		http.Error(w, "Không tìm thấy bài tập", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(ex)
}

// HandleAPIFunctions trả về danh sách hàm tra cứu theo từ khoá
func HandleAPIFunctions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	funcs, err := logic.SearchFunctions(query)
	if err != nil {
		http.Error(w, "Lỗi tìm kiếm hàm", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(funcs)
}
