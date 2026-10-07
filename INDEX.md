# Bản đồ mã nguồn dự án web-python (Programming Learning & Assessment Platform)

Hệ thống học tập, thực hành và đánh giá lập trình trực tuyến toàn diện:
- **Backend**: Sử dụng hoàn toàn thư viện chuẩn Go (`net/http`, `html/template`, `database/sql`, `crypto/*`), không phụ thuộc web framework bên ngoài.
- **Frontend**: Giao diện HTML5, CSS hiện đại và Vanilla JavaScript kết hợp engine Skulpt thực thi Python phía máy khách.
- **Đánh giá tự động**: Tích hợp cả Skulpt runner cho luyện tập tức thì và Server-side Judge chấm bài chính thức qua test case ẩn.
- **Bảo mật & Kiểm toán**: PBKDF2 băm mật khẩu, Session/CSRF an toàn, nhật ký kiểm toán (Audit Trail) và phát hiện gian lận phòng thi.

---

## 1. Cấu trúc thư mục dự án

| Đường dẫn thư mục / tệp | Mô tả chức năng | Vai trò / Kiến trúc |
| :--- | :--- | :--- |
| [`cmd/web/`](file:///mnt/Work/01_Projects/web-python/cmd/web/) | Entry point máy chủ HTTP, cấu hình middleware và ánh xạ toàn bộ tuyến đường (routes). | Điểm khởi chạy ứng dụng (Entry point). |
| [`config/`](file:///mnt/Work/01_Projects/web-python/config/) | Tệp cấu hình môi trường `.env` (`PORT`, `DB_PATH`). | Quản lý cấu hình & biến môi trường. |
| [`data/Database/`](file:///mnt/Work/01_Projects/web-python/data/Database/) | Thư mục lưu trữ tệp cơ sở dữ liệu SQLite (`algo_db.db`). | Tầng lưu trữ dữ liệu bền vững. |
| [`internal/`](file:///mnt/Work/01_Projects/web-python/internal/) | Mã nguồn nghiệp vụ nội bộ (Internal Packages), phân tách theo domain. | Tầng nghiệp vụ cốt lõi (Domain Logic). |
| [`web/templates/`](file:///mnt/Work/01_Projects/web-python/web/templates/) | Các mẫu giao diện HTML (`html/template`) cho Sinh viên, Giảng viên và Quản trị. | Tầng trình bày (Presentation / Views). |
| [`web/static/`](file:///mnt/Work/01_Projects/web-python/web/static/) | Tài nguyên tĩnh (CSS, JavaScript, Engine Skulpt, hình ảnh). | Tài nguyên máy khách (Static Assets). |
| [`ROADMAP.md`](file:///mnt/Work/01_Projects/web-python/ROADMAP.md) | Kế hoạch lộ trình chi tiết 25 giai đoạn chuyển đổi hệ thống. | Kế hoạch kiến trúc dự án. |
| [`README.md`](file:///mnt/Work/01_Projects/web-python/README.md) | Tài liệu hướng dẫn cài đặt, khởi chạy và danh mục tài khoản mẫu. | Hướng dẫn vận hành dự án. |
| [`INDEX.md`](file:///mnt/Work/01_Projects/web-python/INDEX.md) | Bản đồ cấu trúc tổng quan và mục lục mã nguồn toàn dự án. | Mục lục điều hướng mã nguồn. |

---

## 2. Chi tiết các Package nội bộ (`internal/`)

Mỗi package trong `internal/` tuân thủ nguyên tắc thiết kế phân tầng: Repository (truy vấn DB) → Service (nghiệp vụ & kiểm tra logic) → Handler (xử lý HTTP request/response).

| Package | Trách nhiệm chính | File tiêu biểu |
| :--- | :--- | :--- |
| [`internal/auth/`](file:///mnt/Work/01_Projects/web-python/internal/auth/) | Xác thực người dùng, băm mật khẩu PBKDF2, quản lý phiên và phân quyền vai trò (`student`, `teacher`, `admin`). | [`auth.go`](file:///mnt/Work/01_Projects/web-python/internal/auth/auth.go), [`middleware.go`](file:///mnt/Work/01_Projects/web-python/internal/auth/middleware.go) |
| [`internal/security/`](file:///mnt/Work/01_Projects/web-python/internal/security/) | Cơ chế bảo mật: CSRF Token bảo vệ biểu mẫu, làm sạch đầu vào ngăn chặn XSS. | [`csrf.go`](file:///mnt/Work/01_Projects/web-python/internal/security/csrf.go), [`sanitizer.go`](file:///mnt/Work/01_Projects/web-python/internal/security/sanitizer.go) |
| [`internal/audit/`](file:///mnt/Work/01_Projects/web-python/internal/audit/) | Ghi nhật ký kiểm toán hệ thống, middleware tự động ghi nhận request và truy vết thao tác nhạy cảm. | [`audit.go`](file:///mnt/Work/01_Projects/web-python/internal/audit/audit.go) |
| [`internal/database/`](file:///mnt/Work/01_Projects/web-python/internal/database/) | Khởi tạo kết nối SQLite, hệ thống migration tự động (001–015) và nạp dữ liệu mẫu ban đầu (seed). | [`db.go`](file:///mnt/Work/01_Projects/web-python/internal/database/db.go), [`migration.go`](file:///mnt/Work/01_Projects/web-python/internal/database/migration.go), [`seed_courses.go`](file:///mnt/Work/01_Projects/web-python/internal/database/seed_courses.go) |
| [`internal/course/`](file:///mnt/Work/01_Projects/web-python/internal/course/) | Quản lý danh mục khóa học (mã môn, tiêu đề, giảng viên phụ trách, trạng thái xuất bản). | [`course.go`](file:///mnt/Work/01_Projects/web-python/internal/course/course.go) |
| [`internal/class/`](file:///mnt/Work/01_Projects/web-python/internal/class/) | Quản lý lớp học theo học kỳ, ghi danh sinh viên (Enrollment) và danh sách lớp học của tôi. | [`class.go`](file:///mnt/Work/01_Projects/web-python/internal/class/class.go) |
| [`internal/lesson/`](file:///mnt/Work/01_Projects/web-python/internal/lesson/) | Quản lý chương (Chapter) và bài học (Lesson), hiển thị cây chương trình học phân cấp. | [`lesson.go`](file:///mnt/Work/01_Projects/web-python/internal/lesson/lesson.go) |
| [`internal/exercise/`](file:///mnt/Work/01_Projects/web-python/internal/exercise/) | Quản lý ngân hàng bài tập lập trình, cấu hình hàm cho phép (Whitelist) và test case công khai/ẩn. | [`exercise.go`](file:///mnt/Work/01_Projects/web-python/internal/exercise/exercise.go) |
| [`internal/practice/`](file:///mnt/Work/01_Projects/web-python/internal/practice/) | Quản lý bài tập thực hành trên IDE, tự động lưu bản nháp và ghi nhận trạng thái giải bài. | [`practice.go`](file:///mnt/Work/01_Projects/web-python/internal/practice/practice.go) |
| [`internal/submission/`](file:///mnt/Work/01_Projects/web-python/internal/submission/) | Quản lý nộp bài chính thức, đếm số lần thử, lưu lịch sử nộp và kết quả chấm điểm. | [`submission.go`](file:///mnt/Work/01_Projects/web-python/internal/submission/submission.go) |
| [`internal/judge/`](file:///mnt/Work/01_Projects/web-python/internal/judge/) | Trình chấm tự động phía máy chủ: kiểm thử mã Python với test case ẩn, đối soát output chuẩn. | [`judge.go`](file:///mnt/Work/01_Projects/web-python/internal/judge/judge.go) |
| [`internal/assignment/`](file:///mnt/Work/01_Projects/web-python/internal/assignment/) | Quản lý bài tập về nhà, giao bài cho lớp, đặt hạn nộp và tổng hợp điểm số. | [`assignment.go`](file:///mnt/Work/01_Projects/web-python/internal/assignment/assignment.go) |
| [`internal/exam/`](file:///mnt/Work/01_Projects/web-python/internal/exam/) | Quản lý kỳ thi, phòng thi trực tuyến, đếm ngược thời gian từ server và giám sát gian lận (anti-cheat). | [`exam.go`](file:///mnt/Work/01_Projects/web-python/internal/exam/exam.go) |
| [`internal/quiz/`](file:///mnt/Work/01_Projects/web-python/internal/quiz/) | Quản lý câu hỏi trắc nghiệm đính kèm bài học và chấm điểm trắc nghiệm. | [`quiz.go`](file:///mnt/Work/01_Projects/web-python/internal/quiz/quiz.go) |
| [`internal/dashboard/`](file:///mnt/Work/01_Projects/web-python/internal/dashboard/) | Bảng điều khiển sinh viên: tổng hợp tiến độ học, thống kê bài tập hoàn thành và thông báo hạn chót. | [`dashboard.go`](file:///mnt/Work/01_Projects/web-python/internal/dashboard/dashboard.go) |
| [`internal/teacher/`](file:///mnt/Work/01_Projects/web-python/internal/teacher/) | Bảng điều khiển giảng viên: thống kê lớp học, phát hiện sinh viên nguy cơ và quản lý bài nộp. | [`teacher.go`](file:///mnt/Work/01_Projects/web-python/internal/teacher/teacher.go) |
| [`internal/progress/`](file:///mnt/Work/01_Projects/web-python/internal/progress/) | Tính toán và tổng hợp tiến độ hoàn thành theo từng môn, chương và bài học. | [`progress.go`](file:///mnt/Work/01_Projects/web-python/internal/progress/progress.go) |
| [`internal/logic/`](file:///mnt/Work/01_Projects/web-python/internal/logic/) | Lấy cấu trúc khóa học (`GetCourseStructure`), tra cứu hàm cho phép và tương thích ngược `Topics`. | [`exercise.go`](file:///mnt/Work/01_Projects/web-python/internal/logic/exercise.go), [`function.go`](file:///mnt/Work/01_Projects/web-python/internal/logic/function.go) |
| [`internal/frontend/`](file:///mnt/Work/01_Projects/web-python/internal/frontend/) | HTTP Handler phục vụ giao diện IDE và tra cứu hàm API. | [`handlers.go`](file:///mnt/Work/01_Projects/web-python/internal/frontend/handlers.go) |

---

## 3. Cấu trúc Giao diện (`web/templates/`)

- [`base.html`](file:///mnt/Work/01_Projects/web-python/web/templates/base.html): Khung bố cục chung của IDE.
- [`ide.html`](file:///mnt/Work/01_Projects/web-python/web/templates/ide.html): Giao diện lập trình, cây chương mục bài tập, console và runner.
- [`login.html`](file:///mnt/Work/01_Projects/web-python/web/templates/login.html): Trang đăng nhập hệ thống với xác thực bảo mật.
- [`dashboard.html`](file:///mnt/Work/01_Projects/web-python/web/templates/dashboard.html): Trang tổng quan học tập của sinh viên.
- [`courses.html`](file:///mnt/Work/01_Projects/web-python/web/templates/courses.html) & [`course_detail.html`](file:///mnt/Work/01_Projects/web-python/web/templates/course_detail.html): Danh sách và đề cương môn học.
- [`lesson.html`](file:///mnt/Work/01_Projects/web-python/web/templates/lesson.html): Trang bài học, lý thuyết kèm bài tập liên kết.
- [`my_classes.html`](file:///mnt/Work/01_Projects/web-python/web/templates/my_classes.html): Danh sách lớp học đã tham gia.
- [`my_assignments.html`](file:///mnt/Work/01_Projects/web-python/web/templates/my_assignments.html): Danh sách bài tập về nhà của sinh viên.
- [`my_exams.html`](file:///mnt/Work/01_Projects/web-python/web/templates/my_exams.html) & [`exam_take.html`](file:///mnt/Work/01_Projects/web-python/web/templates/exam_take.html): Giao diện phòng thi trực tuyến có giám sát.
- Các màn hình giảng viên: [`teacher_dashboard.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_dashboard.html), [`teacher_classes.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_classes.html), [`teacher_curriculum.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_curriculum.html), [`teacher_assignments.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_assignments.html), [`teacher_exams.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_exams.html), [`teacher_exam_monitoring.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_exam_monitoring.html), [`teacher_submissions.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_submissions.html), [`teacher_class_analytics.html`](file:///mnt/Work/01_Projects/web-python/web/templates/teacher_class_analytics.html).
