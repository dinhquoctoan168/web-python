# Web Python — Programming Learning & Assessment Platform

Nền tảng học tập và đánh giá lập trình trực tuyến toàn diện, chuyển đổi từ mô hình Web Python IDE sang hệ sinh thái quản lý đào tạo, thực hành và kiểm tra thi cử tự động.

---

## 1. Triết lý thiết kế & Công nghệ cốt lõi

- **Backend thuần Go Standard Library**:
  - Không sử dụng các web framework bên thứ ba (Gin, Echo, Fiber).
  - Sử dụng 100% thư viện chuẩn Go: `net/http`, `database/sql`, `html/template`, `crypto/subtle`, `crypto/sha256`, `crypto/rand`.
  - Cơ sở dữ liệu: SQLite nhúng (`modernc.org/sqlite`), hoạt động độc lập không yêu cầu CGO.
- **Frontend tối giản & hiệu năng cao**:
  - Vanilla HTML5, CSS hiện đại (CSS Variables, Flexbox, Grid), Vanilla JavaScript (không dùng React, Vue, Bootstrap, Tailwind).
  - **Skulpt Engine**: Trình thông dịch Python viết bằng JavaScript, thực thi mã nguồn ngay trên trình duyệt máy khách (Client-side Execution) cho bài tập thực hành.
  - **Server-side Judge**: Trình chấm tự động phía máy chủ chạy các bộ test case ẩn (Hidden Tests) phục vụ nộp bài chính thức, bài tập về nhà và bài thi.
- **Bảo mật & Kiểm toán**:
  - Quản lý phiên (Session) an toàn, cơ chế CSRF Token cho toàn bộ tác vụ thay đổi dữ liệu (POST).
  - Băm mật khẩu chuẩn PBKDF2/SHA-256.
  - Hệ thống ghi nhật ký kiểm toán (Audit Trail) cho toàn bộ hành vi của sinh viên và giảng viên.

---

## 2. Tính năng chính

### 2.1. Không gian lập trình (IDE)
- Giao diện lập trình trực quan mô phỏng phong cách hiện đại.
- Danh mục tra cứu cú pháp và hàm chuẩn được phép sử dụng (Function Whitelist).
- Kiểm tra bài làm tức thời với bộ kiểm thử công khai (Public Test Runner).
- Tự động lưu bản nháp mã nguồn và đồng bộ trạng thái thực hành.

### 2.2. Quản lý Đào tạo (Course, Chapter, Lesson)
- Cấu trúc chương trình học phân cấp: **Khóa học (Course) → Chương (Chapter) → Bài học (Lesson) → Bài tập (Exercise)**.
- Phân loại bài tập theo độ khó (Dễ, Trung bình, Khó) và hỗ trợ trắc nghiệm (Quiz).
- Theo dõi tiến độ học tập (Progress Tracking) theo từng môn, từng chương và từng bài học.

### 2.3. Lớp học & Sinh viên (Class & Enrollment)
- Giảng viên tạo lớp học theo học kỳ và ghi danh sinh viên vào từng lớp.
- Sinh viên theo dõi danh sách lớp học đã tham gia và các bài tập được giao.

### 2.4. Bài tập về nhà (Assignment)
- Giảng viên giao bài tập cho từng lớp học với hạn nộp cụ thể và chế độ công bố (Draft / Published).
- Chấm điểm tự động và tổng hợp bảng điểm theo lớp.

### 2.5. Kỳ thi & Đánh giá (Exam & Assessment)
- Tạo đề thi với thời lượng giới hạn, số lần làm bài và mật khẩu phòng thi.
- Bộ đếm thời gian được xác thực an toàn từ phía máy chủ (Server-side Countdown), ngăn chặn sửa đổi thời gian máy khách.
- Giám sát gian lận thời gian thực: Ghi nhận sự kiện chuyển tab, rời màn hình toàn màn hình, copy/paste mã nguồn.
- Màn hình giám sát phòng thi dành riêng cho giám thị/giảng viên.

### 2.6. Bảng điều khiển Giảng viên (Teacher Analytics)
- Báo cáo thống kê tỉ lệ hoàn thành bài tập của sinh viên.
- Cảnh báo sinh viên có nguy cơ tụt hậu học tập (At-risk students).
- Xem lịch sử nộp bài, mã nguồn và kết quả kiểm thử của từng sinh viên.

---

## 3. Cài đặt & Khởi chạy

### Yêu cầu môi trường
- **Go**: Phiên bản 1.22 trở lên.
- **Git**

### Cấu hình biến môi trường
Tệp cấu hình mặc định nằm tại [`config/.env`](file:///mnt/Work/01_Projects/web-python/config/.env):
```env
PORT=8080
DB_PATH=data/Database/algo_db.db
```

### Khởi chạy máy chủ
```bash
# Nạp dependency và chạy máy chủ
go run ./cmd/web
```
Máy chủ sẽ tự động chạy cơ chế di chuyển lược đồ (Migrations 001 → 015) và nạp dữ liệu mẫu nếu cơ sở dữ liệu trống.

Truy cập hệ thống tại: [http://localhost:8080](http://localhost:8080)

### Chạy kiểm thử tự động
```bash
go test -v ./...
```

---

## 4. Tài khoản mẫu (Seed Accounts)

Hệ thống đã tự động tạo sẵn các tài khoản thử nghiệm sau khi khởi động:

| Vai trò | Tên đăng nhập | Mật khẩu | Mục đích |
| :--- | :--- | :--- | :--- |
| **Quản trị viên (Admin)** | `admin` | `admin123` | Quản lý hệ thống toàn diện |
| **Giảng viên (Teacher)** | `teacher` | `teacher123` | Quản lý khóa học, lớp học, đề thi, chấm bài |
| **Sinh viên (Student)** | `student` | `student123` | Học tập, làm bài tập, nộp bài, tham gia thi |

---

## 5. Danh mục môn học mặc định

Hệ thống đi kèm 3 môn học hoàn chỉnh:
1. **PY101 — Nhập môn Lập trình Python**: 10 chương từ cú pháp biến, vòng lặp đến xử lý chuỗi và tệp tin.
2. **PRG201 — Lập trình Hướng đối tượng Python**: 6 chương về Class, Kế thừa, Đa hình và Ngoại lệ.
3. **DSA301 — Cấu trúc Dữ liệu & Giải thuật**: 11 chương bao gồm Mảng, Ngăn xếp, Hàng đợi, Bảng băm, Cây nhị phân và các giải thuật Sắp xếp/Tìm kiếm.

---

## 6. Bản đồ định tuyến (Route Map)

### Tuyến công khai (Public)
- `GET /` — Điều hướng thông minh (theo trạng thái đăng nhập & vai trò).
- `GET /login`, `POST /login` — Trang và API đăng nhập.
- `POST /logout` — Đăng xuất và hủy phiên.

### Tuyến sinh viên (Student)
- `GET /dashboard` — Bảng điều khiển học tập cá nhân.
- `GET /courses`, `GET /course?id=` — Danh sách môn học và chi tiết chương trình.
- `GET /lesson?id=` — Nội dung bài học và bài tập liên quan.
- `GET /ide` — Không gian lập trình IDE.
- `GET /my-classes` — Danh sách lớp học đang tham gia.
- `GET /my-assignments` — Danh sách bài tập được giao.
- `GET /my-exams`, `GET /exam/take?id=` — Danh sách kỳ thi và phòng thi trực tuyến.

### Tuyến giảng viên (Teacher)
- `GET /teacher`, `GET /teacher/dashboard` — Bảng điều khiển giảng viên.
- `GET /teacher/class/analytics` — Thống kê & cảnh báo học tập theo lớp.
- `GET /teacher/courses`, `GET /teacher/course/new` — Quản lý môn học.
- `GET /teacher/classes`, `GET /teacher/class?id=` — Quản lý lớp học và ghi danh.
- `GET /teacher/curriculum?course_id=` — Biên soạn chương và bài học.
- `GET /teacher/assignments`, `GET /teacher/assignment/new` — Quản lý và giao bài tập.
- `GET /teacher/exams`, `GET /teacher/exam/new` — Tạo và công bố đề thi.
- `GET /teacher/exam/monitoring?id=` — Giám sát phòng thi trực tuyến.
- `GET /teacher/submissions` — Quản lý và chấm bài nộp.

### Giao diện lập trình ứng dụng (API)
- `GET /api/me` — Thông tin người dùng hiện tại.
- `GET /api/exercise?id=` — Chi tiết bài tập và public test cases.
- `POST /api/practice/save` — Lưu bản nháp bài làm.
- `POST /api/practice/submit` — Kiểm thử mã nguồn bài thực hành.
- `POST /api/submission` — Nộp bài chính thức (chấm server-side với hidden test cases).
- `POST /api/exam/save-answer` — Lưu tạm câu trả lời bài thi.
- `POST /api/exam/submit` — Nộp bài thi chính thức.
- `POST /api/exam/event` — Ghi nhận sự kiện giám sát gian lận phòng thi.
- `GET /api/progress/course?id=` — Truy vấn tiến độ học tập.
