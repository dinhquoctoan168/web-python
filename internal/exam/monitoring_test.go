package exam

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestExamMonitoring(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	// 1. Tạo và mở đề thi
	nowStr := time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04")
	laterStr := time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04")
	exam, err := svc.CreateExam(2, 1, "Kiểm tra giám sát", "Mô tả", 60, nowStr, laterStr, []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("CreateExam thất bại: %v", err)
	}
	_ = svc.PublishExam(exam.ID)

	// 2. Thí sinh bắt đầu phiên thi
	session, _, _, err := svc.StartOrResumeSession(exam.ID, 3)
	if err != nil {
		t.Fatalf("StartOrResumeSession thất bại: %v", err)
	}

	// 3. Ghi nhận các sự kiện vi phạm
	_ = svc.RecordSessionEvent(session.ID, "tab_hidden", "Rời khỏi tab 12 giây")
	_ = svc.RecordSessionEvent(session.ID, "tab_hidden", "Rời khỏi tab 5 giây")
	_ = svc.RecordSessionEvent(session.ID, "window_blur", "Mất tiêu điểm cửa sổ")
	_ = svc.RecordSessionEvent(session.ID, "paste_attempt", "Dán đoạn code 45 ký tự")
	_ = svc.RecordSessionEvent(session.ID, "fullscreen_exit", "Thoát chế độ toàn màn hình")

	// Thử gửi sự kiện không hợp lệ
	errInvalid := svc.RecordSessionEvent(session.ID, "invalid_hack_event", "")
	if errInvalid == nil {
		t.Errorf("Kỳ vọng lỗi khi gửi loại sự kiện không hợp lệ")
	}

	// 4. Giảng viên kiểm tra báo cáo giám sát
	_, summaries, err := svc.GetMonitoringReport(exam.ID)
	if err != nil {
		t.Fatalf("GetMonitoringReport thất bại: %v", err)
	}

	if len(summaries) != 1 {
		t.Fatalf("Kỳ vọng 1 thí sinh trong báo cáo giám sát, nhận %d", len(summaries))
	}

	s := summaries[0]
	if s.TabHiddenCount != 2 {
		t.Errorf("Kỳ vọng TabHiddenCount = 2, nhận %d", s.TabHiddenCount)
	}
	if s.WindowBlurCount != 1 {
		t.Errorf("Kỳ vọng WindowBlurCount = 1, nhận %d", s.WindowBlurCount)
	}
	if s.PasteAttemptCount != 1 {
		t.Errorf("Kỳ vọng PasteAttemptCount = 1, nhận %d", s.PasteAttemptCount)
	}
	if s.FullscreenExitCount != 1 {
		t.Errorf("Kỳ vọng FullscreenExitCount = 1, nhận %d", s.FullscreenExitCount)
	}
	if s.TotalWarnings != 5 {
		t.Errorf("Kỳ vọng TotalWarnings = 5, nhận %d", s.TotalWarnings)
	}
	if len(s.Events) != 5 {
		t.Errorf("Kỳ vọng 5 sự kiện trong timeline, nhận %d", len(s.Events))
	}
}
