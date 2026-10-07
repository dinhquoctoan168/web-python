package class

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"web_python/internal/auth"
)

func TestClassOwnership(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed teacher 1, teacher 2 (khác chủ), student 1, student 2
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(1, 'teacher1', 'hash', 'Thay Giao 1', 'teacher'),
		(2, 'teacher2', 'hash', 'Thay Giao 2', 'teacher'),
		(3, 'student1', 'hash', 'Sinh Vien 1', 'student'),
		(4, 'student2', 'hash', 'Sinh Vien 2', 'student')`)
	_, _ = db.Exec(`INSERT INTO courses (id, code, name) VALUES (1, 'DSA301', 'Cau truc du lieu')`)

	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc, nil)

	// Tạo lớp thuộc teacher 1
	cl, err := svc.CreateClass(1, "Lop 01", "HK1", "2026-2027", 1)
	if err != nil {
		t.Fatalf("CreateClass lỗi: %v", err)
	}

	// Ghi danh student 1
	if err := svc.EnrollStudent(cl.ID, 3); err != nil {
		t.Fatalf("EnrollStudent lỗi: %v", err)
	}

	// 1. Teacher 2 (khác chủ) truy cập HandleTeacherClassDetail -> 403
	reqDetail := httptest.NewRequest("GET", "/teacher/class?id="+strconv.Itoa(cl.ID), nil)
	reqDetail = reqDetail.WithContext(auth.WithUser(reqDetail.Context(), &auth.User{ID: 2, Role: auth.RoleTeacher}))
	recDetail := httptest.NewRecorder()
	handler.HandleTeacherClassDetail(recDetail, reqDetail)
	if recDetail.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi giảng viên khác xem chi tiết lớp, nhận %d", recDetail.Code)
	}

	// 2. Teacher 2 cố thêm sinh viên 2 vào lớp -> 403
	formEnroll := url.Values{"class_id": {strconv.Itoa(cl.ID)}, "username": {"student2"}}
	reqEnroll := httptest.NewRequest("POST", "/teacher/class/enroll", strings.NewReader(formEnroll.Encode()))
	reqEnroll.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqEnroll = reqEnroll.WithContext(auth.WithUser(reqEnroll.Context(), &auth.User{ID: 2, Role: auth.RoleTeacher}))
	recEnroll := httptest.NewRecorder()
	handler.HandleTeacherEnrollStudent(recEnroll, reqEnroll)
	if recEnroll.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi giảng viên khác enroll student, nhận %d", recEnroll.Code)
	}

	// 3. Teacher 2 cố xóa sinh viên 1 khỏi lớp -> 403
	formRemove := url.Values{"class_id": {strconv.Itoa(cl.ID)}, "student_id": {"3"}}
	reqRemove := httptest.NewRequest("POST", "/teacher/class/remove-student", strings.NewReader(formRemove.Encode()))
	reqRemove.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqRemove = reqRemove.WithContext(auth.WithUser(reqRemove.Context(), &auth.User{ID: 2, Role: auth.RoleTeacher}))
	recRemove := httptest.NewRecorder()
	handler.HandleTeacherRemoveStudent(recRemove, reqRemove)
	if recRemove.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng 403 Forbidden khi giảng viên khác remove student, nhận %d", recRemove.Code)
	}

	// 4. Teacher 1 (chủ sở hữu) truy cập HandleTeacherClassDetail -> 200
	reqOwnerDetail := httptest.NewRequest("GET", "/teacher/class?id="+strconv.Itoa(cl.ID), nil)
	reqOwnerDetail = reqOwnerDetail.WithContext(auth.WithUser(reqOwnerDetail.Context(), &auth.User{ID: 1, Role: auth.RoleTeacher}))
	recOwnerDetail := httptest.NewRecorder()
	handler.HandleTeacherClassDetail(recOwnerDetail, reqOwnerDetail)
	if recOwnerDetail.Code != http.StatusOK {
		t.Errorf("Kỳ vọng 200 OK cho chủ sở hữu lớp, nhận %d", recOwnerDetail.Code)
	}

	// 5. Admin truy cập HandleTeacherClassDetail -> 200
	reqAdminDetail := httptest.NewRequest("GET", "/teacher/class?id="+strconv.Itoa(cl.ID), nil)
	reqAdminDetail = reqAdminDetail.WithContext(auth.WithUser(reqAdminDetail.Context(), &auth.User{ID: 99, Role: auth.RoleAdmin}))
	recAdminDetail := httptest.NewRecorder()
	handler.HandleTeacherClassDetail(recAdminDetail, reqAdminDetail)
	if recAdminDetail.Code != http.StatusOK {
		t.Errorf("Kỳ vọng 200 OK cho admin, nhận %d", recAdminDetail.Code)
	}
}
