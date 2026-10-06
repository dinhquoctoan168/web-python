# Thư mục web

Chứa toàn bộ tầng giao diện (View) và tài nguyên tĩnh của ứng dụng:
- `templates/base.html`: Khung layout HTML chuẩn dùng chung.
- `templates/ide.html`: Giao diện chi tiết IDE (Đề bài, Tra cứu hàm, Trình soạn thảo, Bảng Console).
- `static/css/ide.css`: Định kiểu giao diện phong cách IDE tối màu (Dark Theme), chia khung linh hoạt.
- `static/js/skulpt.min.js` & `skulpt-stdlib.js`: Engine biên dịch Python sang JavaScript chạy thuần máy khách.
- `static/js/engine.js`: Bộ engine kiểm duyệt giới hạn hàm (whitelist) và điều phối Skulpt chạy mã.
- `static/js/ide.js`: Logic điều khiển giao diện IDE (phím tắt, tìm kiếm hàm, chuyển bài tập, tab kết quả).
