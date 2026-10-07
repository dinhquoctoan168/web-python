// Package frontend
// Mục đích: Cung cấp các tiện ích xử lý dữ liệu trang chung và API hàm tra cứu.
package frontend

import (
	"encoding/json"
	"net/http"

	"web_python/internal/logic"
)

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
