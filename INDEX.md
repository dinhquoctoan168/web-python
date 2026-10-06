# Bản đồ mã nguồn dự án web-python (Web Python IDE - CSDL & Giải thuật)

Hệ thống học tập Cơ sở dữ liệu và Giải thuật trực tuyến bằng Python:
- **Backend**: Sử dụng hoàn toàn thư viện chuẩn Go (`net/http`, `html/template`, `database/sql`), không dùng framework.
- **Frontend**: Giao diện IDE mô phỏng phong cách VS Code với HTML, CSS và JavaScript thuần (Vanilla JS), không dùng UI framework.
- **Cơ chế thực thi**: Mã Python do sinh viên gõ được chuyển đổi sang JavaScript và thực thi trực tiếp trên máy khách (client-side) thông qua engine Skulpt.
- **Bảo mật & Giới hạn**: Tích hợp bộ kiểm soát danh mục hàm (Function Whitelist) và bảng tra cứu hàm trực quan.

## Cấu trúc thư mục

| Đường dẫn thư mục / tệp | Mô tả chức năng | Vai trò / Kiến trúc |
| :--- | :--- | :--- |
| `cmd/web/` | Chứa entry point chạy máy chủ HTTP phục vụ website. | Điểm khởi chạy (Entry point). |
| `config/` | Chứa tệp cấu hình môi trường `.env` (cổng mạng, đường dẫn database). | Cấu hình & Môi trường. |
| `internal/database/` | Kết nối SQLite, khởi tạo schema bảng bài tập, chủ đề và danh mục hàm. | Tầng dữ liệu (Database Layer). |
| `internal/logic/` | Nghiệp vụ truy vấn bài tập, lọc chủ đề và tìm kiếm hàm. | Tầng nghiệp vụ (Business Logic). |
| `internal/frontend/` | Bộ điều hướng HTTP handler và render template HTML. | Tầng điều khiển (Handlers / Controller). |
| `web/templates/` | Các mẫu giao diện HTML (`base.html`, `ide.html`). | Giao diện hiển thị (View Layer). |
| `web/static/` | Tài nguyên tĩnh (CSS IDE, JS Controller, Engine Skulpt Python-to-JS). | Tài nguyên máy khách (Static Assets). |
| `data/Database/` | Thư mục lưu tệp cơ sở dữ liệu SQLite (`algo_db.db`). | Lưu trữ dữ liệu SQLite. |
| `INDEX.md` | Bản đồ tổng quan toàn bộ cấu trúc dự án. | Lập mục lục dự án. |
