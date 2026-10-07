// Package ide
// Mục đích: Định nghĩa HTTP handler và render giao diện lập trình IDE cho học viên và giảng viên.
package ide

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"web_python/internal/assignment"
	"web_python/internal/auth"
	"web_python/internal/course"
	"web_python/internal/exercise"
	"web_python/internal/lesson"
	"web_python/internal/logic"
	"web_python/internal/security"
)

// PageData chứa dữ liệu truyền vào template HTML của IDE
type PageData struct {
	Title           string
	CSRFToken       string
	CourseCode      string
	CourseName      string
	Mode            string
	AssignmentID    int
	Chapters        []exercise.ChapterWithExercises
	Topics          []logic.Topic
	CurrentExercise *exercise.ClientExerciseDetail
	Functions       []logic.FunctionItem
}

// Handler điều phối logic hiển thị cho giao diện IDE
type Handler struct {
	courseService     *course.Service
	lessonService     *lesson.Service
	exerciseService   *exercise.Service
	assignmentService *assignment.Service
}

// NewHandler khởi tạo Handler với đầy đủ domain services
func NewHandler(
	cs *course.Service,
	ls *lesson.Service,
	es *exercise.Service,
	as *assignment.Service,
) *Handler {
	return &Handler{
		courseService:     cs,
		lessonService:     ls,
		exerciseService:   es,
		assignmentService: as,
	}
}

// HandleIDE render trang giao diện lập trình IDE sử dụng thống nhất exercise.Service
func (h *Handler) HandleIDE(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/ide" {
		http.NotFound(w, r)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "practice"
	}

	var currentExercise *exercise.ClientExerciseDetail
	var courseCode string
	var courseName string
	var assignmentID int
	var chapters []exercise.ChapterWithExercises

	exIDStr := strings.TrimSpace(r.URL.Query().Get("id"))
	if exIDStr != "" {
		if id, err := strconv.Atoi(exIDStr); err == nil && id > 0 {
			currentExercise, _ = h.exerciseService.GetExerciseForClient(id)
		}
	}

	// 1. Xử lý chế độ Assignment nếu có assignment_id
	if mode == "assignment" {
		assignmentIDStr := strings.TrimSpace(r.URL.Query().Get("assignment_id"))
		if assignmentIDStr == "" {
			http.Error(w, "Thiếu mã đợt bài tập (assignment_id)", http.StatusBadRequest)
			return
		}
		aid, err := strconv.Atoi(assignmentIDStr)
		if err != nil || aid <= 0 {
			http.Error(w, "ID bài tập không hợp lệ", http.StatusBadRequest)
			return
		}
		if h.assignmentService == nil {
			http.Error(w, "Dịch vụ bài tập chưa sẵn sàng", http.StatusInternalServerError)
			return
		}

		user := auth.GetUser(r.Context())
		if user == nil {
			http.Error(w, "Yêu cầu đăng nhập để truy cập bài tập", http.StatusForbidden)
			return
		}

		var asgn *assignment.Assignment
		if user.Role == auth.RoleStudent {
			asgn, err = h.assignmentService.GetAssignmentForStudent(aid, user.ID)
			if err != nil || asgn == nil {
				http.Error(w, "Bạn không có quyền truy cập bài tập này", http.StatusForbidden)
				return
			}
		} else {
			asgn, err = h.assignmentService.GetAssignment(aid)
			if err != nil || asgn == nil {
				http.Error(w, "Không tìm thấy bài tập", http.StatusNotFound)
				return
			}
			if user.Role == auth.RoleTeacher && asgn.CreatedBy != user.ID && user.Role != auth.RoleAdmin {
				http.Error(w, "Bạn không có quyền quản lý bài tập này", http.StatusForbidden)
				return
			}
		}

		assignmentID = aid

		// Nếu có chỉ định exercise id cụ thể, kiểm tra exercise đó có thuộc assignment không
		if exIDStr != "" {
			if id, err := strconv.Atoi(exIDStr); err == nil && id > 0 {
				contains, err := h.assignmentService.AssignmentContainsExercise(aid, id)
				if err != nil || !contains {
					http.Error(w, "Bài tập không thuộc đợt giao này", http.StatusForbidden)
					return
				}
			}
		}

		courseCode = asgn.CourseCode
		courseName = asgn.CourseName
		asgnChapter := exercise.ChapterWithExercises{
			ID:          asgn.ID,
			Title:       asgn.Title,
			Description: asgn.Description,
		}
		for _, ae := range asgn.Exercises {
			if d, err := h.exerciseService.GetExerciseForClient(ae.ExerciseID); err == nil && d != nil {
				asgnChapter.Exercises = append(asgnChapter.Exercises, *d)
			}
		}
		chapters = append(chapters, asgnChapter)
		if currentExercise == nil && len(asgnChapter.Exercises) > 0 {
			currentExercise = &asgnChapter.Exercises[0]
		}
	}

	// 2. Chế độ Practice thông thường
	if len(chapters) == 0 {
		var targetCourse *course.Course
		courseParam := strings.TrimSpace(r.URL.Query().Get("course"))

		// Ưu tiên 1: Lấy course từ exercise nếu đã xác định được exercise
		if currentExercise != nil && currentExercise.CourseID > 0 {
			targetCourse, _ = h.courseService.GetCourseByID(currentExercise.CourseID)
		}

		// Ưu tiên 2: Lấy course theo query param (mã code hoặc ID)
		if targetCourse == nil && courseParam != "" {
			targetCourse, _ = h.courseService.GetCourseByCode(courseParam)
			if targetCourse == nil {
				if cid, err := strconv.Atoi(courseParam); err == nil && cid > 0 {
					targetCourse, _ = h.courseService.GetCourseByID(cid)
				}
			}
		}

		// Ưu tiên 3: Lấy môn học đầu tiên học viên đang theo học / active
		if targetCourse == nil {
			courses, err := h.courseService.ListCoursesForStudent()
			if err == nil && len(courses) > 0 {
				targetCourse = &courses[0]
			}
		}

		if targetCourse != nil {
			courseCode = targetCourse.Code
			courseName = targetCourse.Name
			var err error
			chapters, err = h.exerciseService.GetCourseStructure(targetCourse.ID)
			if err != nil {
				log.Printf("Lỗi lấy cấu trúc bài tập môn %d: %v", targetCourse.ID, err)
			}
		}

		// Mặc định chọn bài đầu tiên nếu chưa chọn
		if currentExercise == nil {
			for _, ch := range chapters {
				if len(ch.Exercises) > 0 {
					currentExercise = &ch.Exercises[0]
					break
				}
			}
		}
	}

	funcs, err := logic.SearchFunctions("")
	if err != nil {
		log.Printf("Lỗi lấy danh sách functions: %v", err)
	}

	title := "Web Python IDE"
	if courseName != "" {
		title = "Web Python IDE - " + courseName
	}

	data := PageData{
		Title:           title,
		CSRFToken:       security.GetTokenFromContext(r.Context()),
		CourseCode:      courseCode,
		CourseName:      courseName,
		Mode:            mode,
		AssignmentID:    assignmentID,
		Chapters:        chapters,
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

// HandleAPIFunctions trả về danh sách hàm tra cứu theo từ khoá
func (h *Handler) HandleAPIFunctions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	funcs, err := logic.SearchFunctions(query)
	if err != nil {
		http.Error(w, "Lỗi tìm kiếm hàm", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(funcs)
}
