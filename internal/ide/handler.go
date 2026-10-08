// Package ide
// Mục đích: Định nghĩa HTTP handler và render giao diện lập trình IDE cho học viên và giảng viên.
package ide

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"web_python/internal/assignment"
	"web_python/internal/auth"
	"web_python/internal/course"
	"web_python/internal/exercise"
	"web_python/internal/frontend"
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
	User            *auth.User
	Nav             frontend.NavigationData
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

	user := auth.GetUser(r.Context())
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "practice"
	}

	var currentExercise *exercise.ClientExerciseDetail
	var asgn *assignment.Assignment
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

		if user == nil {
			http.Error(w, "Yêu cầu đăng nhập để truy cập bài tập", http.StatusForbidden)
			return
		}

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

	var validLesson *lesson.Lesson
	var targetCourse *course.Course

	// 2. Chế độ Practice thông thường
	if len(chapters) == 0 {
		courseParam := strings.TrimSpace(r.URL.Query().Get("course"))
		if courseParam == "" {
			courseParam = strings.TrimSpace(r.URL.Query().Get("course_id"))
		}

		// Xác thực bài học lý thuyết nếu có tham số lesson_id
		lessonIDStr := strings.TrimSpace(r.URL.Query().Get("lesson_id"))
		if lid, err := strconv.Atoi(lessonIDStr); err == nil && lid > 0 && h.lessonService != nil {
			isTeacher := user != nil && (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin)
			if l, _, err := h.lessonService.GetLessonDetail(lid, isTeacher); err == nil && l != nil {
				validLesson = l
			}
		}

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

		// Ưu tiên 3: Lấy course từ lesson đã xác thực
		if targetCourse == nil && validLesson != nil && validLesson.CourseID > 0 {
			targetCourse, _ = h.courseService.GetCourseByID(validLesson.CourseID)
		}

		// Ưu tiên 4: Lấy môn học đầu tiên học viên đang theo học / active
		if targetCourse == nil {
			courses, err := h.courseService.ListCoursesForStudent()
			if err == nil && len(courses) > 0 {
				targetCourse = &courses[0]
			}
		}

		// Kiểm tra xung đột: nếu lesson không thuộc targetCourse đang luyện tập thì bỏ qua lesson context phụ
		if validLesson != nil && targetCourse != nil && validLesson.CourseID != targetCourse.ID {
			validLesson = nil
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

	hasExplicitCourse := r.URL.Query().Get("course") != "" || r.URL.Query().Get("course_id") != ""
	ideCtx := BuildValidatedIDEContext(user, mode, asgn, validLesson, targetCourse, hasExplicitCourse, currentExercise)
	csrfToken := security.GetTokenFromContext(r.Context())
	nav := frontend.BuildNavigationData(user, "ide", ideCtx.Breadcrumbs, ideCtx.BackURL, ideCtx.BackLabel, csrfToken)

	data := PageData{
		Title:           title,
		CSRFToken:       csrfToken,
		CourseCode:      courseCode,
		CourseName:      courseName,
		Mode:            mode,
		AssignmentID:    assignmentID,
		Chapters:        chapters,
		CurrentExercise: currentExercise,
		Functions:       funcs,
		User:            user,
		Nav:             nav,
	}

	baseFile := frontend.ResolveTemplatePath(filepath.Join("web", "templates", "base.html"))
	ideFile := frontend.ResolveTemplatePath(filepath.Join("web", "templates", "ide.html"))
	allFiles := []string{baseFile, ideFile}
	allFiles = append(allFiles, frontend.GetSharedTemplatePaths()...)

	tmpl, err := template.ParseFiles(allFiles...)
	if err != nil {
		log.Printf("Lỗi parse template: %v", err)
		http.Error(w, "Lỗi hiển thị giao diện", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, filepath.Base(baseFile), data); err != nil {
		log.Printf("Lỗi execute template: %v", err)
	}
}

// IDEBackContext chứa thông tin điều hướng và breadcrumb đã được xác thực
type IDEBackContext struct {
	BackURL     string
	BackLabel   string
	Breadcrumbs []frontend.Breadcrumb
}

// BuildValidatedIDEContext tạo BackURL, BackLabel và Breadcrumbs dựa trên các thực thể đã xác thực
func BuildValidatedIDEContext(user *auth.User, mode string, asgn *assignment.Assignment, validLesson *lesson.Lesson, targetCourse *course.Course, hasExplicitCourse bool, currentExercise *exercise.ClientExerciseDetail) IDEBackContext {
	isTeacher := user != nil && (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin)
	var exTitle string
	if currentExercise != nil && currentExercise.Title != "" {
		exTitle = currentExercise.Title
	} else {
		exTitle = "IDE"
	}

	// 1. Ngữ cảnh Assignment (đã được xác thực trong handler)
	if mode == "assignment" && asgn != nil {
		if isTeacher {
			return IDEBackContext{
				BackURL:   fmt.Sprintf("/teacher/assignment?id=%d", asgn.ID),
				BackLabel: "Quay lại Bài tập",
				Breadcrumbs: []frontend.Breadcrumb{
					{Label: "Teacher", URL: "/teacher"},
					{Label: "Bài tập", URL: "/teacher/assignments"},
					{Label: asgn.Title, URL: fmt.Sprintf("/teacher/assignment?id=%d", asgn.ID)},
					{Label: exTitle, URL: ""},
				},
			}
		}
		return IDEBackContext{
			BackURL:   fmt.Sprintf("/assignment?id=%d", asgn.ID),
			BackLabel: "Quay lại Bài tập",
			Breadcrumbs: []frontend.Breadcrumb{
				{Label: "Trang chủ", URL: "/dashboard"},
				{Label: "Bài tập", URL: "/assignments"},
				{Label: asgn.Title, URL: fmt.Sprintf("/assignment?id=%d", asgn.ID)},
				{Label: exTitle, URL: ""},
			},
		}
	}

	// 2. Ngữ cảnh Lesson (đã xác thực qua lessonService và thuộc đúng course)
	if validLesson != nil {
		if isTeacher {
			bc := []frontend.Breadcrumb{
				{Label: "Teacher", URL: "/teacher"},
				{Label: "Môn học", URL: "/teacher/courses"},
			}
			if targetCourse != nil {
				bc = append(bc, frontend.Breadcrumb{
					Label: targetCourse.Name,
					URL:   fmt.Sprintf("/teacher/curriculum?course_id=%d", targetCourse.ID),
				})
			}
			bc = append(bc, frontend.Breadcrumb{
				Label: validLesson.Title,
				URL:   fmt.Sprintf("/lesson?id=%d", validLesson.ID),
			})
			bc = append(bc, frontend.Breadcrumb{
				Label: "IDE",
				URL:   "",
			})
			return IDEBackContext{
				BackURL:     fmt.Sprintf("/lesson?id=%d", validLesson.ID),
				BackLabel:   "Quay lại Bài học",
				Breadcrumbs: bc,
			}
		}
		bc := []frontend.Breadcrumb{
			{Label: "Trang chủ", URL: "/dashboard"},
			{Label: "Khóa học", URL: "/courses"},
		}
		if targetCourse != nil {
			bc = append(bc, frontend.Breadcrumb{
				Label: targetCourse.Name,
				URL:   fmt.Sprintf("/course?id=%d", targetCourse.ID),
			})
		}
		bc = append(bc, frontend.Breadcrumb{
			Label: validLesson.Title,
			URL:   fmt.Sprintf("/lesson?id=%d", validLesson.ID),
		})
		bc = append(bc, frontend.Breadcrumb{
			Label: "IDE",
			URL:   "",
		})
		return IDEBackContext{
			BackURL:     fmt.Sprintf("/lesson?id=%d", validLesson.ID),
			BackLabel:   "Quay lại Bài học",
			Breadcrumbs: bc,
		}
	}

	// 3. Ngữ cảnh Course rõ ràng (có course param trong query)
	if hasExplicitCourse && targetCourse != nil {
		if isTeacher {
			return IDEBackContext{
				BackURL:   fmt.Sprintf("/teacher/curriculum?course_id=%d", targetCourse.ID),
				BackLabel: "Quay lại Chương trình",
				Breadcrumbs: []frontend.Breadcrumb{
					{Label: "Teacher", URL: "/teacher"},
					{Label: "Môn học", URL: "/teacher/courses"},
					{Label: targetCourse.Name, URL: fmt.Sprintf("/teacher/curriculum?course_id=%d", targetCourse.ID)},
					{Label: exTitle, URL: ""},
				},
			}
		}
		return IDEBackContext{
			BackURL:   fmt.Sprintf("/course?id=%d", targetCourse.ID),
			BackLabel: "Quay lại Môn học",
			Breadcrumbs: []frontend.Breadcrumb{
				{Label: "Trang chủ", URL: "/dashboard"},
				{Label: "Khóa học", URL: "/courses"},
				{Label: targetCourse.Name, URL: fmt.Sprintf("/course?id=%d", targetCourse.ID)},
				{Label: exTitle, URL: ""},
			},
		}
	}

	// 4. Mở trực tiếp tự do
	if isTeacher {
		return IDEBackContext{
			BackURL:   "/teacher",
			BackLabel: "Quay lại Bảng điều khiển",
			Breadcrumbs: []frontend.Breadcrumb{
				{Label: "Teacher", URL: "/teacher"},
				{Label: "IDE", URL: ""},
			},
		}
	}
	if user != nil {
		return IDEBackContext{
			BackURL:   "/dashboard",
			BackLabel: "Quay lại Trang chủ",
			Breadcrumbs: []frontend.Breadcrumb{
				{Label: "Trang chủ", URL: "/dashboard"},
				{Label: "IDE", URL: ""},
			},
		}
	}
	return IDEBackContext{
		BackURL:   "/courses",
		BackLabel: "Quay lại Môn học",
		Breadcrumbs: []frontend.Breadcrumb{
			{Label: "Khóa học", URL: "/courses"},
			{Label: "IDE", URL: ""},
		},
	}
}

// ResolveIDEBackContext xác định BackURL và BackLabel theo tham số query và vai trò người dùng (loại bỏ exam giả lập)
func ResolveIDEBackContext(r *http.Request, user *auth.User) (string, string) {
	isTeacher := user != nil && (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin)

	// 1. Ngữ cảnh làm bài tập (assignment)
	assignmentIDStr := strings.TrimSpace(r.URL.Query().Get("assignment_id"))
	if aid, err := strconv.Atoi(assignmentIDStr); err == nil && aid > 0 {
		if isTeacher {
			return fmt.Sprintf("/teacher/assignment?id=%d", aid), "Quay lại Bài tập"
		}
		return fmt.Sprintf("/assignment?id=%d", aid), "Quay lại Bài tập"
	}

	// 2. Ngữ cảnh từ bài học lý thuyết
	lessonIDStr := strings.TrimSpace(r.URL.Query().Get("lesson_id"))
	if lid, err := strconv.Atoi(lessonIDStr); err == nil && lid > 0 {
		return fmt.Sprintf("/lesson?id=%d", lid), "Quay lại Bài học"
	}

	// 3. Ngữ cảnh từ môn học cụ thể
	courseIDStr := strings.TrimSpace(r.URL.Query().Get("course_id"))
	if cid, err := strconv.Atoi(courseIDStr); err == nil && cid > 0 {
		if isTeacher {
			return fmt.Sprintf("/teacher/curriculum?course_id=%d", cid), "Quay lại Chương trình"
		}
		return fmt.Sprintf("/course?id=%d", cid), "Quay lại Môn học"
	}

	courseCode := strings.TrimSpace(r.URL.Query().Get("course"))
	if courseCode != "" {
		if isTeacher {
			return "/teacher/courses", "Quay lại Môn học"
		}
		return "/courses", "Quay lại Môn học"
	}

	// 4. Mặc định theo vai trò người dùng
	if isTeacher {
		return "/teacher", "Quay lại Bảng điều khiển"
	}
	if user != nil {
		return "/dashboard", "Quay lại Trang chủ"
	}
	return "/courses", "Quay lại Môn học"
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
