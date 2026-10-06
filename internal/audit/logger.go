package audit

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// LogRequestError ghi nhận lỗi phát sinh khi xử lý HTTP request (4xx, 5xx)
// Tuyệt đối không log thông tin cookie phiên hay password
func LogRequestError(r *http.Request, statusCode int, err error) {
	if r == nil {
		log.Printf("[REQUEST ERROR] status=%d err=%v", statusCode, err)
		return
	}
	path := r.URL.Path
	method := r.Method
	remoteAddr := r.RemoteAddr
	log.Printf("[REQUEST ERROR] method=%s path=%s remote=%s status=%d err=%v", method, path, remoteAddr, statusCode, err)
}

// LogDatabaseError ghi nhận lỗi tương tác với cơ sở dữ liệu
func LogDatabaseError(operation string, err error) {
	if err == nil {
		return
	}
	log.Printf("[DATABASE ERROR] operation=%s err=%v", operation, err)
}

// LogLoginFailure ghi nhận đăng nhập thất bại (chỉ log username và IP, KHÔNG log password)
func LogLoginFailure(username string, ip string, reason string) {
	username = strings.TrimSpace(username)
	if username == "" {
		username = "<empty>"
	}
	log.Printf("[AUTH FAILURE] Đăng nhập thất bại: user=%s ip=%s reason=%s", username, ip, reason)
}

// LogSubmissionFailure ghi nhận lỗi trong quá trình nộp bài
func LogSubmissionFailure(studentID int, exerciseID int, reason string) {
	log.Printf("[SUBMISSION FAILURE] Nộp bài thất bại: student_id=%d exercise_id=%d reason=%s", studentID, exerciseID, reason)
}

// LogJudgeError ghi nhận lỗi phát sinh trong hệ thống chấm bài tự động
func LogJudgeError(submissionID int, exerciseID int, err error) {
	log.Printf("[JUDGE ERROR] Lỗi chấm bài: submission_id=%d exercise_id=%d err=%v", submissionID, exerciseID, err)
}

// responseWriterRecorder bọc http.ResponseWriter để bắt mã trạng thái HTTP phục vụ ghi log
type responseWriterRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseWriterRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// RequestLoggerMiddleware middleware tự động log các request và cảnh báo khi có lỗi 4xx/5xx
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseWriterRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		if rec.statusCode >= 400 {
			log.Printf("[HTTP %d] %s %s (%v)", rec.statusCode, r.Method, r.URL.Path, duration)
		}
	})
}
