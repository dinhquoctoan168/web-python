package frontend

import (
	"net/http/httptest"
	"testing"
)

func TestBasePageData(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	data := NewBasePageData(req, "Tiêu đề kiểm thử")

	if data.Title != "Tiêu đề kiểm thử" {
		t.Errorf("Mong đợi Title 'Tiêu đề kiểm thử', nhận: %s", data.Title)
	}

	m := map[string]any{"Key": "Value"}
	injected := InjectCSRFToMap(req, m)
	if _, ok := injected["CSRFToken"]; !ok {
		t.Errorf("Mong đợi CSRFToken được inject vào map")
	}
}
