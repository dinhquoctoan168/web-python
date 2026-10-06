# Thư mục internal

Chứa toàn bộ mã nguồn tầng logic, database và handlers:
- `database/db.go`: Quản lý SQLite, tự động tạo bảng (topics, exercises, functions) và nạp dữ liệu mẫu ban đầu.
- `logic/exercise.go`: Logic nghiệp vụ truy vấn bài tập, danh sách chủ đề và tìm kiếm hàm theo từ khóa.
- `frontend/handlers.go`: Bộ điều hướng HTTP handler (`net/http`) và render template HTML qua `html/template`.
