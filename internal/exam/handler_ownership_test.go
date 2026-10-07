package exam

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"web_python/internal/auth"
)

func TestExamOwnership(t *testing.T) {
	db, examSvc := setupTestDB(t)
	defer db.Close()

	// Thêm 1 giảng viên khác (id 4)
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(4, 'teacher2', 'hash', 'Thay Giao Khac', 'teacher')`)

	handler := NewHandler(examSvc, nil, nil)

	// Tạo bài thi bởi teacher 1 (id 2)
	now := time.Now()
	startStr := now.Add(-1 * time.Hour).Format("2006-01-02T15:04")
	endStr := now.Add(2 * time.Hour).Format("2006-01-02T15:04")
	ex, err := examSvc.CreateExam(2, 1, "Kiểm tra giữa kỳ", "Mô tả", 60, startStr, endStr, []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("CreateExam thất bại: %v", err)
	}

	// 1. Giảng viên khác (id 4) cố truy cập HandleTeacherExamMonitoring -> 403
	reqMon := httptest.NewRequest("GET", "/teacher/exam/monitoring?id="+strconv.Itoa(ex.ID), nil)
	reqMon = reqMon.WithContext(auth.WithUser(reqMon.Context(), &auth.User{ID: 4, Role: auth.RoleTeacher}))
	recMon := httptest.NewRecorder()
	handler.HandleTeacherExamMonitoring(recMon, reqMon)
	if recMon.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi giảng viên khác xem monitoring, nhận %d", recMon.Code)
	}

	// 2. Giảng viên khác cố publish đề thi -> 403
	formPub := url.Values{"id": {strconv.Itoa(ex.ID)}}
	reqPub := httptest.NewRequest("POST", "/teacher/exam/publish", strings.NewReader(formPub.Encode()))
	reqPub.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPub = reqPub.WithContext(auth.WithUser(reqPub.Context(), &auth.User{ID: 4, Role: auth.RoleTeacher}))
	recPub := httptest.NewRecorder()
	handler.HandleTeacherPublishExam(recPub, reqPub)
	if recPub.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi giảng viên khác publish bài thi, nhận %d", recPub.Code)
	}

	// 3. Giảng viên chính chủ (id 2) truy cập monitoring -> 200
	reqOwnerMon := httptest.NewRequest("GET", "/teacher/exam/monitoring?id="+strconv.Itoa(ex.ID), nil)
	reqOwnerMon = reqOwnerMon.WithContext(auth.WithUser(reqOwnerMon.Context(), &auth.User{ID: 2, Role: auth.RoleTeacher}))
	recOwnerMon := httptest.NewRecorder()
	handler.HandleTeacherExamMonitoring(recOwnerMon, reqOwnerMon)
	if recOwnerMon.Code != http.StatusOK {
		t.Errorf("Kỳ vọng 200 OK cho chính chủ khi xem monitoring, nhận %d", recOwnerMon.Code)
	}

	// 4. Giảng viên chính chủ (id 2) publish đề thi -> 303
	reqOwnerPub := httptest.NewRequest("POST", "/teacher/exam/publish", strings.NewReader(formPub.Encode()))
	reqOwnerPub.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqOwnerPub = reqOwnerPub.WithContext(auth.WithUser(reqOwnerPub.Context(), &auth.User{ID: 2, Role: auth.RoleTeacher}))
	recOwnerPub := httptest.NewRecorder()
	handler.HandleTeacherPublishExam(recOwnerPub, reqOwnerPub)
	if recOwnerPub.Code != http.StatusSeeOther {
		t.Errorf("Kỳ vọng 303 SeeOther khi chính chủ publish bài thi, nhận %d", recOwnerPub.Code)
	}

	// 5. Admin truy cập monitoring -> 200
	reqAdminMon := httptest.NewRequest("GET", "/teacher/exam/monitoring?id="+strconv.Itoa(ex.ID), nil)
	reqAdminMon = reqAdminMon.WithContext(auth.WithUser(reqAdminMon.Context(), &auth.User{ID: 99, Role: auth.RoleAdmin}))
	recAdminMon := httptest.NewRecorder()
	handler.HandleTeacherExamMonitoring(recAdminMon, reqAdminMon)
	if recAdminMon.Code != http.StatusOK {
		t.Errorf("Kỳ vọng 200 OK cho admin khi xem monitoring, nhận %d", recAdminMon.Code)
	}
}
