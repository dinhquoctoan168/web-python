// Package main
// Mục đích: Khởi chạy máy chủ HTTP thuần thư viện chuẩn Go, nạp cấu hình và định tuyến URL.
package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"web_python/internal/assignment"
	"web_python/internal/auth"
	"web_python/internal/class"
	"web_python/internal/course"
	"web_python/internal/dashboard"
	"web_python/internal/database"
	"web_python/internal/exam"
	"web_python/internal/exercise"
	"web_python/internal/frontend"
	"web_python/internal/judge"
	"web_python/internal/lesson"
	"web_python/internal/practice"
	"web_python/internal/progress"
	"web_python/internal/quiz"
	"web_python/internal/security"
	"web_python/internal/submission"
	"web_python/internal/teacher"
)

// loadEnv đọc tệp cấu hình .env thủ công bằng standard library để không phụ thuộc lib ngoài
func loadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Thông báo: Không tìm thấy tệp env %s: %v", filepath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
}

func main() {
	log.Println("==================================================")
	log.Println("  Hệ thống học tập Web Python IDE (CSDL & Giải thuật) ")
	log.Println("==================================================")

	// 1. Nạp biến môi trường
	loadEnv("config/.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/Database/algo_db.db"
	}

	// 2. Khởi tạo cơ sở dữ liệu
	_, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Khởi tạo database thất bại: %v", err)
	}
	defer database.CloseDB()

	// 3. Khởi tạo các module
	db := database.GetDB()
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService, "web/templates/login.html")
	authMiddleware := auth.NewMiddleware(authService)

	courseRepo := course.NewRepository(db)
	courseService := course.NewService(courseRepo)
	courseHandler := course.NewHandler(courseService)

	classRepo := class.NewRepository(db)
	classService := class.NewService(classRepo)
	classHandler := class.NewHandler(classService, courseService)

	lessonRepo := lesson.NewRepository(db)
	lessonService := lesson.NewService(lessonRepo)
	lessonHandler := lesson.NewHandler(lessonService)

	exerciseRepo := exercise.NewRepository(db)
	exerciseService := exercise.NewService(exerciseRepo)
	exerciseHandler := exercise.NewHandler(exerciseService)

	practiceRepo := practice.NewRepository(db)
	practiceService := practice.NewService(practiceRepo)
	practiceHandler := practice.NewHandler(practiceService)

	submissionRepo := submission.NewRepository(db)
	submissionService := submission.NewService(submissionRepo)

	judgeService := judge.NewService(exerciseService)
	submissionService.SetJudgeService(judgeService)

	submissionHandler := submission.NewHandler(submissionService)

	assignmentRepo := assignment.NewRepository(db)
	assignmentService := assignment.NewService(assignmentRepo)
	assignmentHandler := assignment.NewHandler(assignmentService, classService, exerciseService)

	examRepo := exam.NewRepository(db)
	examService := exam.NewService(examRepo, judgeService)
	examHandler := exam.NewHandler(examService, classService, exerciseService)

	quizRepo := quiz.NewRepository(db)
	quizService := quiz.NewService(quizRepo, practiceService, submissionService)
	quizHandler := quiz.NewHandler(quizService)

	dashboardRepo := dashboard.NewRepository(db)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardService)

	teacherRepo := teacher.NewRepository(db)
	teacherService := teacher.NewService(teacherRepo, db)
	teacherDashboardHandler := teacher.NewHandler(teacherService)

	progressRepo := progress.NewRepository(db)
	progressService := progress.NewService(progressRepo)
	progressHandler := progress.NewHandler(progressService)

	// Liên kết lấy chương trình học vào trang chi tiết môn học
	courseHandler.SetCurriculumFetcher(func(courseID int, isTeacher bool) (any, error) {
		return lessonService.GetCurriculum(courseID, isTeacher)
	})

	// 4. Thiết lập Mux định tuyến thuần standard library
	mux := http.NewServeMux()

	// Phục vụ tài nguyên tĩnh (CSS, JS, Fonts)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Tuyến đường xác thực
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.HandleLogin(w, r)
		} else {
			authHandler.ShowLoginPage(w, r)
		}
	})
	mux.HandleFunc("/logout", authHandler.HandleLogout)
	mux.HandleFunc("/api/me", authHandler.HandleCurrentUser)

	// Tuyến đường môn học, bài học & thi cử (Sinh viên)
	mux.HandleFunc("/dashboard", auth.RequireLogin(dashboardHandler.HandleStudentDashboard))
	mux.HandleFunc("/courses", auth.RequireLogin(courseHandler.HandleListCourses))
	mux.HandleFunc("/course", auth.RequireLogin(courseHandler.HandleCourseDetail))
	mux.HandleFunc("/lesson", auth.RequireLogin(lessonHandler.HandleStudentLesson))
	mux.HandleFunc("/api/course/curriculum", auth.RequireLogin(lessonHandler.HandleAPICourseCurriculum))
	mux.HandleFunc("/my-classes", auth.RequireLogin(classHandler.HandleStudentMyClasses))
	mux.HandleFunc("/my-assignments", auth.RequireLogin(assignmentHandler.HandleStudentMyAssignments))
	mux.HandleFunc("/my-exams", auth.RequireLogin(examHandler.HandleStudentMyExams))
	mux.HandleFunc("/exam/take", auth.RequireLogin(examHandler.HandleStudentTakeExam))

	// Tuyến đường quản lý môn học, lớp học & chương trình giảng dạy (Giảng viên)
	mux.HandleFunc("/teacher", auth.RequireTeacher(teacherDashboardHandler.HandleTeacherDashboard))
	mux.HandleFunc("/teacher/dashboard", auth.RequireTeacher(teacherDashboardHandler.HandleTeacherDashboard))
	mux.HandleFunc("/teacher/class/analytics", auth.RequireTeacher(teacherDashboardHandler.HandleClassAnalytics))
	mux.HandleFunc("/teacher/student/detail", auth.RequireTeacher(teacherDashboardHandler.HandleStudentDetail))
	mux.HandleFunc("/teacher/courses", auth.RequireTeacher(courseHandler.HandleTeacherListCourses))
	mux.HandleFunc("/teacher/course/new", auth.RequireTeacher(courseHandler.HandleTeacherNewCourseForm))
	mux.HandleFunc("/teacher/course/create", auth.RequireTeacher(courseHandler.HandleTeacherCreateCourse))
	mux.HandleFunc("/teacher/course/edit", auth.RequireTeacher(courseHandler.HandleTeacherEditCourseForm))
	mux.HandleFunc("/teacher/course/update", auth.RequireTeacher(courseHandler.HandleTeacherUpdateCourse))

	mux.HandleFunc("/teacher/classes", auth.RequireTeacher(classHandler.HandleTeacherListClasses))
	mux.HandleFunc("/teacher/class", auth.RequireTeacher(classHandler.HandleTeacherClassDetail))
	mux.HandleFunc("/teacher/class/create", auth.RequireTeacher(classHandler.HandleTeacherCreateClass))
	mux.HandleFunc("/teacher/class/enroll", auth.RequireTeacher(classHandler.HandleTeacherEnrollStudent))
	mux.HandleFunc("/teacher/class/remove-student", auth.RequireTeacher(classHandler.HandleTeacherRemoveStudent))

	mux.HandleFunc("/teacher/assignments", auth.RequireTeacher(assignmentHandler.HandleTeacherListAssignments))
	mux.HandleFunc("/teacher/assignment/new", auth.RequireTeacher(assignmentHandler.HandleTeacherNewAssignmentForm))
	mux.HandleFunc("/teacher/assignment/create", auth.RequireTeacher(assignmentHandler.HandleTeacherCreateAssignment))
	mux.HandleFunc("/teacher/assignment", auth.RequireTeacher(assignmentHandler.HandleTeacherAssignmentDetail))
	mux.HandleFunc("/teacher/assignment/publish", auth.RequireTeacher(assignmentHandler.HandleTeacherPublishAssignment))

	mux.HandleFunc("/teacher/exams", auth.RequireTeacher(examHandler.HandleTeacherListExams))
	mux.HandleFunc("/teacher/exam/new", auth.RequireTeacher(examHandler.HandleTeacherNewExamForm))
	mux.HandleFunc("/teacher/exam/create", auth.RequireTeacher(examHandler.HandleTeacherCreateExam))
	mux.HandleFunc("/teacher/exam/publish", auth.RequireTeacher(examHandler.HandleTeacherPublishExam))
	mux.HandleFunc("/teacher/exam/monitoring", auth.RequireTeacher(examHandler.HandleTeacherExamMonitoring))

	mux.HandleFunc("/teacher/curriculum", auth.RequireTeacher(lessonHandler.HandleTeacherCurriculum))
	mux.HandleFunc("/teacher/chapter/create", auth.RequireTeacher(lessonHandler.HandleTeacherCreateChapter))
	mux.HandleFunc("/teacher/lesson/new", auth.RequireTeacher(lessonHandler.HandleTeacherNewLessonForm))
	mux.HandleFunc("/teacher/lesson/create", auth.RequireTeacher(lessonHandler.HandleTeacherCreateLesson))
	mux.HandleFunc("/teacher/lesson/edit", auth.RequireTeacher(lessonHandler.HandleTeacherEditLessonForm))
	mux.HandleFunc("/teacher/lesson/update", auth.RequireTeacher(lessonHandler.HandleTeacherUpdateLesson))

	mux.HandleFunc("/teacher/submissions", auth.RequireTeacher(submissionHandler.HandleTeacherSubmissions))
	mux.HandleFunc("/teacher/submission/view", auth.RequireTeacher(submissionHandler.HandleTeacherSubmissionView))

	// Trang giao diện IDE (yêu cầu đăng nhập)
	mux.HandleFunc("/ide", auth.RequireLogin(frontend.HandleIDE))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		user := auth.GetUser(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if user.Role == "student" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/teacher", http.StatusSeeOther)
	})

	// API endpoints
	mux.HandleFunc("/api/exercise", exerciseHandler.HandleAPIExercise)
	mux.HandleFunc("/api/functions", frontend.HandleAPIFunctions)
	mux.HandleFunc("/api/practice/save", practiceHandler.HandleSaveDraft)
	mux.HandleFunc("/api/practice/submit", practiceHandler.HandleSubmitPractice)
	mux.HandleFunc("/api/practice/state", practiceHandler.HandleGetState)
	mux.HandleFunc("/api/practice/all-states", practiceHandler.HandleGetAllStates)
	mux.HandleFunc("/api/attempt/action", auth.RequireLogin(submissionHandler.HandleRecordAction))
	mux.HandleFunc("/api/submissions", auth.RequireLogin(submissionHandler.HandleCreateSubmission))
	mux.HandleFunc("/api/submission", auth.RequireLogin(submissionHandler.HandleCreateSubmission))
	mux.HandleFunc("/api/submissions/my", auth.RequireLogin(submissionHandler.HandleGetMySubmissions))
	mux.HandleFunc("/api/exam/save-answer", auth.RequireLogin(examHandler.HandleAPISaveAnswer))
	mux.HandleFunc("/api/exam/submit", auth.RequireLogin(examHandler.HandleAPISubmitExam))
	mux.HandleFunc("/api/exam/event", auth.RequireLogin(examHandler.HandleAPIRecordExamEvent))
	mux.HandleFunc("/api/quiz", auth.RequireLogin(quizHandler.HandleGetOptions))
	mux.HandleFunc("/api/quiz/submit", auth.RequireLogin(quizHandler.HandleSubmitQuiz))
	mux.HandleFunc("/api/progress/course", auth.RequireLogin(progressHandler.HandleGetCourseProgress))
	mux.HandleFunc("/api/progress/chapter", auth.RequireLogin(progressHandler.HandleGetChapterProgress))
	mux.HandleFunc("/api/progress/lesson", auth.RequireLogin(progressHandler.HandleGetLessonProgress))

	// Bọc toàn bộ handler với CSRFMiddleware và AuthenticateMiddleware (Phase 19)
	csrfMiddleware := security.CSRFMiddleware("/login", "/logout")
	rootHandler := authMiddleware.AuthenticateMiddleware(csrfMiddleware(mux))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      rootHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Xử lý dừng máy chủ an toàn (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server đang chạy tại http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi máy chủ HTTP: %v", err)
		}
	}()

	<-stop
	log.Println("\nĐang tắt máy chủ an toàn...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Lỗi khi shutdown: %v", err)
	}
	log.Println("Máy chủ đã dừng hoàn tất.")
}
