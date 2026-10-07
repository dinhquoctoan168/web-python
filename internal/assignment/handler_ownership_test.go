package assignment

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"web_python/internal/auth"
	"web_python/internal/class"
)

func TestAssignmentOwnershipAndStudentAccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Thêm 1 giảng viên khác và 1 sinh viên ngoài lớp
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(4, 'teacher2', 'hash', 'Thay Giao Khac', 'teacher'),
		(5, 'student2', 'hash', 'Sinh Vien Ngoai Lop', 'student')`)

	assignmentRepo := NewRepository(db)
	assignmentSvc := NewService(assignmentRepo)

	classRepo := class.NewRepository(db)
	classSvc := class.NewService(classRepo)

	handler := NewHandler(assignmentSvc, classSvc, nil)

	// Tạo assignment bởi teacher 2
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")
	a, err := assignmentSvc.CreateAssignment(2, 1, "BT Python", "Mo ta", "", tomorrow, []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("Tạo assignment thất bại: %v", err)
	}

	// 1. Giảng viên khác (id 4) cố truy cập HandleTeacherAssignmentDetail -> 403
	req := httptest.NewRequest("GET", "/teacher/assignment?id="+strconv.Itoa(a.ID), nil)
	req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 4, Role: auth.RoleTeacher}))
	rec := httptest.NewRecorder()
	handler.HandleTeacherAssignmentDetail(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden cho giảng viên khác, nhận %d", rec.Code)
	}

	// 2. Giảng viên khác cố publish -> 403
	form := url.Values{"id": {strconv.Itoa(a.ID)}}
	reqPubOther := httptest.NewRequest("POST", "/teacher/assignment/publish", strings.NewReader(form.Encode()))
	reqPubOther.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPubOther = reqPubOther.WithContext(auth.WithUser(reqPubOther.Context(), &auth.User{ID: 4, Role: auth.RoleTeacher}))
	recPubOther := httptest.NewRecorder()
	handler.HandleTeacherPublishAssignment(recPubOther, reqPubOther)
	if recPubOther.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi giảng viên khác publish, nhận %d", recPubOther.Code)
	}

	// 3. Học viên thuộc lớp (id 3) truy cập khi bài tập còn DRAFT -> 403
	reqStuDraft := httptest.NewRequest("GET", "/assignment?id="+strconv.Itoa(a.ID), nil)
	reqStuDraft = reqStuDraft.WithContext(auth.WithUser(reqStuDraft.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
	recStuDraft := httptest.NewRecorder()
	handler.HandleStudentAssignmentDetail(recStuDraft, reqStuDraft)
	if recStuDraft.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi bài tập còn nháp, nhận %d", recStuDraft.Code)
	}

	// 4. Giảng viên chính chủ (id 2) publish thành công -> 303
	reqPubOwner := httptest.NewRequest("POST", "/teacher/assignment/publish", strings.NewReader(form.Encode()))
	reqPubOwner.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPubOwner = reqPubOwner.WithContext(auth.WithUser(reqPubOwner.Context(), &auth.User{ID: 2, Role: auth.RoleTeacher}))
	recPubOwner := httptest.NewRecorder()
	handler.HandleTeacherPublishAssignment(recPubOwner, reqPubOwner)
	if recPubOwner.Code != http.StatusSeeOther {
		t.Errorf("Kỳ vọng 303 SeeOther khi chính chủ publish, nhận %d", recPubOwner.Code)
	}

	// 5. Học viên không thuộc lớp (id 5) truy cập sau khi publish -> 403
	reqStuOther := httptest.NewRequest("GET", "/assignment?id="+strconv.Itoa(a.ID), nil)
	reqStuOther = reqStuOther.WithContext(auth.WithUser(reqStuOther.Context(), &auth.User{ID: 5, Role: auth.RoleStudent}))
	recStuOther := httptest.NewRecorder()
	handler.HandleStudentAssignmentDetail(recStuOther, reqStuOther)
	if recStuOther.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi học viên không thuộc lớp, nhận %d", recStuOther.Code)
	}

	// 6. Học viên thuộc lớp (id 3) truy cập sau khi publish -> 200
	reqStuEnrolled := httptest.NewRequest("GET", "/assignment?id="+strconv.Itoa(a.ID), nil)
	reqStuEnrolled = reqStuEnrolled.WithContext(auth.WithUser(reqStuEnrolled.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
	recStuEnrolled := httptest.NewRecorder()
	handler.HandleStudentAssignmentDetail(recStuEnrolled, reqStuEnrolled)
	if recStuEnrolled.Code != http.StatusOK {
		t.Errorf("Kỳ vọng 200 OK khi học viên thuộc lớp truy cập bài tập published, nhận %d: %s", recStuEnrolled.Code, recStuEnrolled.Body.String())
	}
}
