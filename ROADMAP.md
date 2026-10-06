# ROADMAP PHÁT TRIỂN `web-python`
## Programming Learning & Assessment Platform

> Repository hiện tại: `dinhquoctoan168/web-python`  
> Mục tiêu: chuyển Web Python IDE hiện tại thành nền tảng học tập, luyện tập và kiểm tra online cho các học phần:
>
> - Python cơ bản
> - Kỹ thuật lập trình
> - Cấu trúc dữ liệu
> - Giải thuật

---

# 1. Nguyên tắc kỹ thuật bắt buộc

Dự án **không sử dụng framework**.

## Backend

Chỉ sử dụng:

- Go standard library
  - `net/http`
  - `html/template`
  - `database/sql`
  - `encoding/json`
  - `crypto/*`
  - `time`
  - `context`
  - `log`
  - `os`
  - `sync`
- SQLite
- Driver SQLite hiện tại của dự án

Không sử dụng:

- Gin
- Echo
- Fiber
- Chi
- GORM
- Ent
- ORM khác

## Frontend

Chỉ sử dụng:

- HTML
- CSS
- Vanilla JavaScript
- Fetch API
- DOM API
- Web Storage API khi cần
- Fullscreen API
- Page Visibility API
- Skulpt cho Python chạy phía client

Không sử dụng:

- React
- Vue
- Angular
- Svelte
- jQuery
- Bootstrap
- Tailwind
- UI framework khác

## Kiến trúc chung

Giữ kiến trúc nhẹ:

```text
Browser
   |
   | HTTP / JSON
   v
Go net/http
   |
   v
Business Logic
   |
   v
database/sql
   |
   v
SQLite
```

Python luyện tập:

```text
Browser
   |
   v
Skulpt
   |
   v
Run / Public Tests
```

Python chấm chính thức:

```text
Browser
   |
   | Submit code
   v
Go Server
   |
   v
Judge
   |
   v
Hidden Tests
   |
   v
Score
```

---

# 2. Trạng thái dự án hiện tại

Dự án hiện đã có:

```text
cmd/web/
internal/database/
internal/logic/
internal/frontend/

web/templates/
web/static/css/
web/static/js/
```

Các chức năng đã có:

- IDE Python trên trình duyệt
- Editor
- Syntax highlighting
- Console
- Chạy Python bằng Skulpt
- Function whitelist
- Danh sách topic
- Danh sách exercise
- Public test case
- Chấm thử
- Function reference
- Exam timer
- Fullscreen exam mode
- Phát hiện blur / rời tab
- SQLite
- Go `net/http`

Các bảng hiện tại:

```text
topics
exercises
functions
```

Vấn đề kiến trúc hiện tại:

```text
Topic
  |
  └── Exercise
```

Cấu trúc này chưa đủ cho hệ thống học tập nhiều môn.

Cần chuyển thành:

```text
Course
  |
  ├── Chapter
  |     |
  |     └── Lesson
  |            |
  |            └── Exercise
  |
  └── Class
        |
        ├── Students
        ├── Assignments
        └── Exams
```

---

# 3. Mục tiêu kiến trúc cuối

## Student

Sinh viên có thể:

- đăng nhập
- xem các môn được đăng ký
- xem nội dung bài học
- chạy code ví dụ
- làm bài tập lập trình
- làm quiz
- theo dõi tiến độ
- xem bài tập được giao
- làm bài kiểm tra
- nộp code
- xem điểm
- xem lịch sử làm bài

## Teacher

Giảng viên có thể:

- quản lý môn học
- tạo chương
- tạo bài học
- tạo bài tập
- tạo test case
- tạo lớp học
- thêm sinh viên
- giao bài
- tạo bài kiểm tra
- xem bài nộp
- xem điểm
- xem tiến độ sinh viên
- xem thống kê lớp

## Admin

Admin có thể:

- quản lý user
- quản lý course
- quản lý role
- xem trạng thái hệ thống

---

# 4. Cấu trúc thư mục mục tiêu

```text
web-python/
│
├── cmd/
│   └── web/
│       └── main.go
│
├── config/
│   └── .env
│
├── data/
│   └── Database/
│       └── algo_db.db
│
├── internal/
│   │
│   ├── database/
│   │   ├── db.go
│   │   ├── schema.go
│   │   ├── migration.go
│   │   └── seed.go
│   │
│   ├── auth/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   ├── handler.go
│   │   └── middleware.go
│   │
│   ├── course/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── class/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── lesson/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── exercise/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── assignment/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── exam/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── submission/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── progress/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   └── judge/
│       ├── judge.go
│       ├── runner.go
│       └── result.go
│
├── web/
│   │
│   ├── templates/
│   │   ├── base.html
│   │   ├── login.html
│   │   ├── dashboard.html
│   │   │
│   │   ├── course/
│   │   │   ├── list.html
│   │   │   ├── detail.html
│   │   │   └── lesson.html
│   │   │
│   │   ├── exercise/
│   │   │   └── ide.html
│   │   │
│   │   ├── assignment/
│   │   │   ├── list.html
│   │   │   └── detail.html
│   │   │
│   │   ├── exam/
│   │   │   ├── list.html
│   │   │   ├── exam.html
│   │   │   └── result.html
│   │   │
│   │   └── teacher/
│   │       ├── dashboard.html
│   │       ├── courses.html
│   │       ├── classes.html
│   │       ├── students.html
│   │       ├── exercises.html
│   │       └── exams.html
│   │
│   └── static/
│       ├── css/
│       │   ├── base.css
│       │   ├── dashboard.css
│       │   ├── course.css
│       │   ├── ide.css
│       │   └── teacher.css
│       │
│       └── js/
│           ├── app.js
│           ├── auth.js
│           ├── course.js
│           ├── lesson.js
│           ├── ide.js
│           ├── engine.js
│           ├── exam.js
│           └── teacher.js
│
├── go.mod
├── go.sum
└── README.md
```

---

# 5. Nguyên tắc chia module Go

Mỗi module nên có tối đa 4 lớp.

```text
model.go
repository.go
service.go
handler.go
```

Vai trò:

```text
handler
   |
   v
service
   |
   v
repository
   |
   v
database
```

Không truy vấn SQL trực tiếp trong handler.

Không đưa logic HTTP vào repository.

Không đưa HTML vào service.

---

# PHASE 0 — ỔN ĐỊNH SOURCE HIỆN TẠI

## Mục tiêu

Đảm bảo bản hiện tại chạy ổn trước khi mở rộng.

## Nhiệm vụ

### 0.1 Kiểm tra startup

Kiểm tra:

```bash
go mod tidy
go run ./cmd/web
```

Xác nhận:

```text
/
 /ide
 /api/exercise
 /api/functions
```

hoạt động.

### 0.2 Tách seed khỏi `db.go`

Hiện `db.go` đang chứa quá nhiều trách nhiệm.

Tạo:

```text
internal/database/
    db.go
    schema.go
    seed.go
```

`db.go`:

```text
InitDB
GetDB
CloseDB
```

`schema.go`:

```text
initSchema
```

`seed.go`:

```text
seedInitialData
seedFunctions
```

### 0.3 Không thay đổi behavior IDE

Sau refactor phải bảo đảm:

- bài tập vẫn hiển thị
- Run vẫn chạy
- Test vẫn chạy
- whitelist vẫn hoạt động
- function search vẫn hoạt động

## Definition of Done

- source build thành công
- không thay đổi UI
- không làm hỏng database hiện tại

---

# PHASE 1 — DATABASE VERSIONING VÀ MIGRATION

## Mục tiêu

Không tiếp tục tạo schema bằng một hàm lớn duy nhất.

## Bảng mới

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Cấu trúc migration

```text
internal/database/migration.go
```

Định nghĩa:

```go
type Migration struct {
    Version int
    Name    string
    Up      func(*sql.DB) error
}
```

Danh sách:

```text
001_initial
002_users
003_courses
004_classes
005_lessons
006_exercises
007_assignments
008_submissions
009_exams
010_progress
```

## Nhiệm vụ

- tạo bảng `schema_migrations`
- đọc version hiện tại
- chạy migration chưa áp dụng
- mỗi migration chạy transaction
- lỗi migration phải rollback
- log version đã chạy

## Definition of Done

Database cũ có thể nâng cấp mà không xóa dữ liệu.

---

# PHASE 2 — AUTHENTICATION VÀ USER

## Mục tiêu

Thêm hệ thống user tối thiểu.

## Database

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    email TEXT,
    role TEXT NOT NULL,
    is_active INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

Role:

```text
admin
teacher
student
```

## Session

Tạo:

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
```

## Không lưu password plaintext

Hash password bằng cơ chế an toàn.

Không tự viết thuật toán hash.

## Module

```text
internal/auth/
```

### model.go

```go
type User struct
type Session struct
```

### repository.go

```text
FindUserByUsername
FindUserByID
CreateSession
FindSession
DeleteSession
```

### service.go

```text
Authenticate
Login
Logout
GetCurrentUser
```

### middleware.go

```text
RequireLogin
RequireTeacher
RequireAdmin
```

## Route

```text
GET  /login
POST /login
POST /logout
```

## UI

Tạo:

```text
web/templates/login.html
web/static/css/auth.css
```

## Cookie

Session ID lưu trong cookie.

Yêu cầu:

```text
HttpOnly
SameSite=Lax
Secure khi chạy HTTPS
```

## Definition of Done

- user đăng nhập được
- logout được
- student không vào trang teacher
- teacher không vào trang admin

---

# PHASE 3 — COURSE

## Mục tiêu

Tạo cấu trúc môn học.

## Database

```sql
CREATE TABLE courses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'active',
    created_by INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(created_by) REFERENCES users(id)
);
```

Ví dụ:

```text
PY101  Python cơ bản
PRG201 Kỹ thuật lập trình
DSA301 Cấu trúc dữ liệu và giải thuật
```

## Module

```text
internal/course/
```

## Chức năng teacher

- tạo course
- sửa course
- ẩn course
- xem danh sách course

## Route

```text
GET  /teacher/courses
GET  /teacher/course/new
POST /teacher/course/create
GET  /teacher/course/edit?id=
POST /teacher/course/update
```

## Student

```text
GET /courses
GET /course?id=
```

## Definition of Done

Teacher tạo được ít nhất 3 course mẫu.

---

# PHASE 4 — CLASS VÀ ENROLLMENT

## Mục tiêu

Phân biệt môn học và lớp học.

## Database

```sql
CREATE TABLE classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    semester TEXT,
    academic_year TEXT,
    teacher_id INTEGER NOT NULL,
    status TEXT DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(course_id) REFERENCES courses(id),
    FOREIGN KEY(teacher_id) REFERENCES users(id)
);
```

```sql
CREATE TABLE enrollments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    enrolled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(class_id, student_id),
    FOREIGN KEY(class_id) REFERENCES classes(id),
    FOREIGN KEY(student_id) REFERENCES users(id)
);
```

## Ví dụ

```text
Course:
DSA301

Classes:
23CNTT1 - HK1 2026
23CNTT2 - HK1 2026
```

## Teacher UI

```text
Lớp học
  |
  ├── Danh sách sinh viên
  ├── Thêm sinh viên
  ├── Xóa sinh viên
  ├── Assignment
  └── Exam
```

## Route

```text
GET  /teacher/classes
GET  /teacher/class?id=
POST /teacher/class/create

POST /teacher/class/enroll
POST /teacher/class/remove-student
```

## Definition of Done

Một sinh viên có thể được thêm vào nhiều lớp.

---

# PHASE 5 — CHAPTER VÀ LESSON

## Mục tiêu

Chuyển từ Topic sang cấu trúc dạy học thực sự.

## Database

```sql
CREATE TABLE chapters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    order_num INTEGER DEFAULT 0,
    FOREIGN KEY(course_id) REFERENCES courses(id)
);
```

```sql
CREATE TABLE lessons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chapter_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content_html TEXT,
    order_num INTEGER DEFAULT 0,
    is_published INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(chapter_id) REFERENCES chapters(id)
);
```

## Nội dung lesson

Không cần CMS phức tạp.

Cho phép teacher nhập:

- tiêu đề
- đoạn văn
- code block
- heading
- note
- warning

Có thể lưu HTML đã sanitize theo whitelist tag.

## UI student

```text
Course
 |
 ├── Chapter 1
 |    ├── Lesson 1
 |    ├── Lesson 2
 |    └── Exercise
 |
 └── Chapter 2
```

## Trang lesson

Bố cục:

```text
Sidebar chapter     Nội dung bài học

Chapter 1           Title
  Lesson 1          Text
  Lesson 2          Code example
  Exercise          Run example
```

## Code example

Nhúng editor nhỏ:

```text
textarea
Run
Output
```

dùng `PythonEngine.runPython`.

## Definition of Done

Student có thể:

- mở course
- chọn chapter
- đọc lesson
- chạy code ví dụ

---

# PHASE 6 — REFACTOR EXERCISE

## Mục tiêu

Biến exercise hiện tại thành ngân hàng bài tập dùng chung.

## Database

Thay bảng exercise bằng cấu trúc:

```sql
CREATE TABLE exercises (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id INTEGER NOT NULL,
    lesson_id INTEGER,
    title TEXT NOT NULL,
    exercise_type TEXT NOT NULL DEFAULT 'coding',
    difficulty TEXT DEFAULT 'easy',
    description TEXT NOT NULL,
    initial_code TEXT,
    solution_code TEXT,
    solution_hint TEXT,
    allowed_functions TEXT,
    time_limit_ms INTEGER DEFAULT 5000,
    status TEXT DEFAULT 'active',
    created_by INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(course_id) REFERENCES courses(id),
    FOREIGN KEY(lesson_id) REFERENCES lessons(id)
);
```

## Exercise type

```text
coding
multiple_choice
code_tracing
short_answer
```

Ban đầu ưu tiên:

```text
coding
multiple_choice
code_tracing
```

## Test cases

Không tiếp tục lưu tất cả test case dưới JSON trong exercise.

Tạo:

```sql
CREATE TABLE exercise_test_cases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exercise_id INTEGER NOT NULL,
    input_data TEXT,
    call_expression TEXT,
    expected_output TEXT,
    is_hidden INTEGER DEFAULT 0,
    weight REAL DEFAULT 1,
    order_num INTEGER DEFAULT 0,
    FOREIGN KEY(exercise_id) REFERENCES exercises(id)
);
```

## Public tests

```text
is_hidden = 0
```

được phép gửi browser.

## Hidden tests

```text
is_hidden = 1
```

không bao giờ gửi client.

## API

```text
GET /api/exercise?id=
```

chỉ trả:

- exercise
- public tests
- whitelist
- description
- initial code

Không trả:

- hidden tests
- solution
- đáp án

## Definition of Done

IDE hiện tại chạy với schema mới.

---

# PHASE 7 — STUDENT PRACTICE

## Mục tiêu

Tạo chế độ luyện tập hoàn chỉnh.

## Các trạng thái

```text
not_started
in_progress
completed
```

## Database

```sql
CREATE TABLE student_exercise_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    best_score REAL DEFAULT 0,
    attempts INTEGER DEFAULT 0,
    first_started_at DATETIME,
    completed_at DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id, exercise_id)
);
```

## UI

Sidebar:

```text
✓ completed
● in progress
○ not started
```

## Chức năng

- Run
- Public Test
- Reset
- Hint
- Save code
- Submit practice

## Auto Save

JavaScript debounce:

```text
student gõ code
     |
     | 2–3 giây không gõ
     v
POST /api/practice/save
```

Không save mỗi keypress.

## API

```text
POST /api/practice/save
POST /api/practice/submit
GET  /api/practice/state
```

## Definition of Done

Reload browser không mất code đang làm.

---

# PHASE 8 — ATTEMPT VÀ SUBMISSION

## Mục tiêu

Lưu lịch sử học thực tế.

## Database

```sql
CREATE TABLE submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    source_code TEXT NOT NULL,
    score REAL DEFAULT 0,
    passed_tests INTEGER DEFAULT 0,
    total_tests INTEGER DEFAULT 0,
    status TEXT,
    submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(student_id) REFERENCES users(id),
    FOREIGN KEY(exercise_id) REFERENCES exercises(id)
);
```

```sql
CREATE TABLE exercise_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    run_count INTEGER DEFAULT 0,
    test_count INTEGER DEFAULT 0,
    hint_count INTEGER DEFAULT 0,
    FOREIGN KEY(student_id) REFERENCES users(id),
    FOREIGN KEY(exercise_id) REFERENCES exercises(id)
);
```

## Khi Run

Tăng:

```text
run_count
```

## Khi Test

Tăng:

```text
test_count
```

## Khi mở Hint

Tăng:

```text
hint_count
```

## Giá trị dữ liệu

Có thể phân tích:

```text
student A
Run 2
Test 1
Hint 0

student B
Run 21
Test 8
Hint 3
```

## Definition of Done

Teacher xem được lịch sử làm bài từng student.

---

# PHASE 9 — ASSIGNMENT

## Mục tiêu

Exercise là ngân hàng câu hỏi.

Assignment chỉ tham chiếu exercise.

## Database

```sql
CREATE TABLE assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    start_at DATETIME,
    due_at DATETIME,
    status TEXT DEFAULT 'draft',
    created_by INTEGER,
    FOREIGN KEY(class_id) REFERENCES classes(id)
);
```

```sql
CREATE TABLE assignment_exercises (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    points REAL DEFAULT 1,
    order_num INTEGER DEFAULT 0,
    FOREIGN KEY(assignment_id) REFERENCES assignments(id),
    FOREIGN KEY(exercise_id) REFERENCES exercises(id)
);
```

## Teacher workflow

```text
Create Assignment
      |
      v
Select Class
      |
      v
Select Exercises
      |
      v
Set Points
      |
      v
Set Deadline
      |
      v
Publish
```

## Student dashboard

Hiển thị:

```text
Bài tập sắp đến hạn

Stack Practice
Due: 12/10

Sorting Lab
Due: 15/10
```

## Definition of Done

Một exercise có thể nằm trong nhiều assignment.

---

# PHASE 10 — SERVER-SIDE JUDGE

## Mục tiêu

Không dùng test case client cho điểm chính thức.

## Quy tắc bảo mật

Browser không được biết:

```text
hidden test
expected hidden output
solution code
```

## Flow

```text
Student
  |
  | source code
  v
POST /api/submission
  |
  v
Go Backend
  |
  ├── check user
  ├── check assignment
  ├── check deadline
  ├── load hidden tests
  |
  v
Judge
  |
  v
Score
```

## Phase đầu

Có thể giữ Skulpt cho:

```text
Run
Public Test
```

Judge chính thức phải tách module:

```text
internal/judge/
```

## Interface

```go
type JudgeRequest struct {
    ExerciseID int
    SourceCode string
}

type TestResult struct {
    Passed   bool
    Expected string
    Actual   string
    Runtime  int64
}

type JudgeResult struct {
    Score       float64
    PassedTests int
    TotalTests  int
    Tests       []TestResult
}
```

## Lưu ý rất quan trọng

Không chạy code sinh viên trực tiếp bằng:

```go
exec.Command("python", ...)
```

trong cùng quyền user với web server mà không có sandbox.

Nếu chưa có sandbox an toàn:

- chưa triển khai chấm Python server-side thực sự
- hoặc chỉ dùng cho môi trường local/thử nghiệm
- production cần isolation

Các hướng isolation về sau:

```text
Docker
nsjail
Firecracker
sandbox service riêng
```

Nhưng phần web architecture vẫn chuẩn bị interface Judge ngay từ đầu.

## Definition of Done

Browser không nhận hidden tests.

---

# PHASE 11 — EXAM

## Mục tiêu

Tạo bài kiểm tra độc lập với exercise.

## Database

```sql
CREATE TABLE exams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    duration_minutes INTEGER NOT NULL,
    start_at DATETIME,
    end_at DATETIME,
    status TEXT DEFAULT 'draft',
    created_by INTEGER,
    FOREIGN KEY(class_id) REFERENCES classes(id)
);
```

```sql
CREATE TABLE exam_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exam_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    points REAL NOT NULL,
    order_num INTEGER DEFAULT 0,
    FOREIGN KEY(exam_id) REFERENCES exams(id),
    FOREIGN KEY(exercise_id) REFERENCES exercises(id)
);
```

```sql
CREATE TABLE exam_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exam_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    started_at DATETIME,
    submitted_at DATETIME,
    status TEXT,
    final_score REAL DEFAULT 0,
    FOREIGN KEY(exam_id) REFERENCES exams(id),
    FOREIGN KEY(student_id) REFERENCES users(id)
);
```

## Exam UI

Không hiển thị:

```text
hint
solution
hidden tests
expected hidden outputs
```

Chỉ hiển thị:

```text
question
editor
run
submit
timer
```

## Timer

Server là nguồn thời gian chính.

Không dựa hoàn toàn vào:

```js
setInterval()
```

Mỗi lần load:

```text
remaining =
exam_session_deadline - server_time
```

## Auto submit

Khi hết giờ:

```text
client gửi submit
```

Nhưng server vẫn phải kiểm tra deadline.

Nếu request đến trễ:

```text
server quyết định
```

không tin timer client.

---

# PHASE 12 — EXAM MONITORING

## Mục tiêu

Theo dõi hành vi bất thường.

Không tự động đình chỉ chỉ vì một lần blur.

## Database

```sql
CREATE TABLE exam_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    event_data TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(session_id) REFERENCES exam_sessions(id)
);
```

## Event

```text
tab_hidden
window_blur
fullscreen_exit
copy_attempt
paste_attempt
context_menu
devtool_shortcut
```

## JavaScript

Dùng:

```text
visibilitychange
blur
fullscreenchange
copy
paste
keydown
```

## API

```text
POST /api/exam/event
```

## Teacher view

```text
Student A
tab_hidden      2
window_blur     4
fullscreen_exit 0
paste_attempt   1
```

## Chính sách

Không kết luận gian lận chỉ dựa trên một signal.

Dữ liệu dùng để:

- cảnh báo
- xem xét
- đối chiếu

---

# PHASE 13 — QUIZ VÀ CODE TRACING

## Multiple Choice

Database:

```sql
CREATE TABLE exercise_options (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exercise_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    is_correct INTEGER DEFAULT 0,
    order_num INTEGER DEFAULT 0,
    FOREIGN KEY(exercise_id) REFERENCES exercises(id)
);
```

API student không gửi:

```text
is_correct
```

## Code tracing

Ví dụ:

```python
x = 1

for i in range(3):
    x *= 2

print(x)
```

Question:

```text
Chương trình in ra gì?
```

Có thể dùng:

```text
multiple_choice
```

hoặc:

```text
short_answer
```

## Definition of Done

Một lesson có thể chứa cả coding và quiz.

---

# PHASE 14 — STUDENT DASHBOARD

## Mục tiêu

Không mở IDE ngay khi login.

## Trang

```text
/dashboard
```

## Hiển thị

```text
Xin chào Nguyễn Văn A

Môn học

Python cơ bản
Progress 72%

Kỹ thuật lập trình
Progress 54%

CTDL & Giải thuật
Progress 41%
```

## Upcoming

```text
Assignment
Exam
Deadline
```

## Recent results

```text
Exercise
Score
Date
```

## API

Không nhất thiết phải SPA.

Go có thể render server-side toàn bộ dashboard.

JavaScript chỉ dùng cho:

- tương tác
- filter
- async refresh nếu cần

---

# PHASE 15 — TEACHER DASHBOARD

## Trang

```text
/teacher
```

## Hiển thị

```text
Course
Classes
Students
Assignments
Exams
```

## Class analytics

Ví dụ:

```text
23CNTT1

Students: 42

Chapter 1 completion: 95%
Chapter 2 completion: 76%
Chapter 3 completion: 48%
```

## Student table

```text
Name
Progress
Average score
Incomplete exercises
Last activity
```

## Student detail

```text
Exercise
Attempts
Run
Test
Hint
Score
Time
```

## Definition of Done

Teacher xác định được sinh viên đang gặp khó ở bài nào.

---

# PHASE 16 — PROGRESS ENGINE

## Mục tiêu

Tính tiến độ nhất quán.

## Rule đơn giản

Lesson complete nếu:

```text
required exercises complete
```

Chapter:

```text
completed lessons / total lessons
```

Course:

```text
completed required activities / total required activities
```

## Không lưu mọi phần trăm

Ưu tiên tính từ dữ liệu nguồn.

Chỉ cache nếu cần performance.

## API

```text
GET /api/progress/course?id=
GET /api/progress/chapter?id=
```

---

# PHASE 17 — ALGORITHM VISUALIZATION

Không phải MVP.

Chỉ làm sau khi core LMS ổn.

## Visualization type

Thêm field:

```text
visualization_type
```

Ví dụ:

```text
array
stack
queue
binary_search
sorting
tree
graph
```

## JavaScript

```text
web/static/js/visualization/
```

Ví dụ:

```text
array.js
stack.js
binary-search.js
sorting.js
```

Không dùng visualization framework.

Dùng:

- HTML
- CSS
- SVG
- Vanilla JS

## Binary Search

State:

```js
{
    array: [],
    left: 0,
    right: 0,
    mid: 0,
    step: 0
}
```

Control:

```text
Previous
Next
Reset
Auto Play
```

---

# PHASE 18 — UI/UX RESTRUCTURE

## Student navigation

```text
Dashboard
Môn học
Bài tập
Kiểm tra
Kết quả
```

## Course navigation

```text
Tổng quan
Nội dung
Bài tập
Kiểm tra
Điểm
```

## Teacher navigation

```text
Dashboard
Môn học
Lớp học
Nội dung
Ngân hàng bài tập
Bài tập
Kiểm tra
Sinh viên
Thống kê
```

## Responsive

Không cần mobile IDE hoàn hảo ngay.

Ưu tiên:

```text
Desktop
Laptop
Tablet
```

IDE nên ưu tiên:

```text
>= 1024px
```

---

# PHASE 19 — SECURITY CƠ BẢN

## SQL

Luôn dùng parameterized query:

```go
db.Query(
    "SELECT ... WHERE id = ?",
    id,
)
```

Không nối chuỗi SQL bằng input người dùng.

## XSS

Với:

```text
html/template
```

giữ escaping mặc định.

Không dùng:

```go
template.HTML
```

trừ khi content đã sanitize.

## CSRF

Các action thay đổi dữ liệu:

```text
POST
```

không dùng GET.

Thêm CSRF token cho:

```text
create
update
delete
submit
```

## Session

Cookie:

```text
HttpOnly
SameSite=Lax
Secure production
```

## Authorization

Không chỉ ẩn button.

Server luôn kiểm tra:

```text
teacher có sở hữu class không
student có thuộc class không
student có quyền làm exam không
```

---

# PHASE 20 — LOGGING

## Log server

Ghi:

```text
request error
database error
login failure
submission failure
judge error
```

Không log:

```text
password
session token
```

## Audit log

Có thể thêm:

```sql
CREATE TABLE audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action TEXT,
    object_type TEXT,
    object_id INTEGER,
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

Dùng cho:

```text
create exam
edit score
delete student
publish assignment
```

---

# PHASE 21 — TESTING

Không dùng framework web nhưng vẫn dùng Go testing.

## Go

```text
*_test.go
```

Test:

```text
repository
service
handler
```

Dùng:

```go
net/http/httptest
```

## Các test quan trọng

### Auth

```text
login success
login fail
expired session
role denied
```

### Course

```text
teacher create
student cannot create
```

### Exercise

```text
hidden tests never returned
```

### Exam

```text
student cannot start early
student cannot submit after deadline
timer calculated by server
```

---

# PHASE 22 — SEED DỮ LIỆU HỌC TẬP

Tạo seed riêng.

```text
internal/database/seed_courses.go
```

Dữ liệu mẫu:

```text
Python cơ bản
Kỹ thuật lập trình
CTDL & Giải thuật
```

## Python

```text
Biến
Kiểu dữ liệu
if
for
while
function
list
dict
file
exception
```

## Kỹ thuật lập trình

```text
function design
module
error handling
OOP
testing
debugging
```

## DSA

```text
array
stack
queue
linked list
tree
graph
sorting
searching
recursion
greedy
dynamic programming
```

---

# PHASE 23 — DATA MIGRATION TỪ SOURCE HIỆN TẠI

Các `topics` hiện tại có thể map vào:

```text
Course:
CTDL & Giải thuật

Topic:
Mảng & Danh sách
Ngăn xếp & Hàng đợi
Bảng băm
Cây
Sorting
```

Chuyển thành:

```text
chapters
```

Các exercise hiện tại giữ lại.

Migration:

```text
topics
   |
   v
chapters

exercises.topic_id
   |
   v
exercises.lesson_id / course_id
```

Không xóa dữ liệu cũ ngay.

Quy trình:

```text
1. backup database
2. create new tables
3. migrate
4. verify counts
5. run application
6. remove legacy schema ở phase sau
```

---

# PHASE 24 — DEPRECATE LEGACY TOPIC MODEL

Chỉ làm khi hệ thống mới ổn định.

Loại dần:

```text
topics
topic_id
GetTopicsWithExercises
```

Thay bằng:

```text
courses
chapters
lessons
GetCourseStructure
```

---

# 6. ROUTE MAP CUỐI

## Public

```text
GET  /
GET  /login
POST /login
POST /logout
```

## Student

```text
GET /dashboard

GET /courses
GET /course?id=
GET /lesson?id=

GET /exercise?id=
GET /assignments
GET /assignment?id=

GET /exams
GET /exam?id=
GET /results
```

## Student API

```text
GET  /api/exercise?id=
POST /api/practice/save
POST /api/practice/submit

POST /api/exam/start
POST /api/exam/save
POST /api/exam/submit
POST /api/exam/event

GET /api/progress/course?id=
```

## Teacher

```text
GET /teacher

GET /teacher/courses
GET /teacher/course

GET /teacher/classes
GET /teacher/class

GET /teacher/lessons

GET /teacher/exercises
GET /teacher/exercise

GET /teacher/assignments
GET /teacher/assignment

GET /teacher/exams
GET /teacher/exam

GET /teacher/students
GET /teacher/student
```

---

# 7. QUY TẮC API

JSON response thống nhất:

Success:

```json
{
  "ok": true,
  "data": {}
}
```

Error:

```json
{
  "ok": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Dữ liệu không hợp lệ"
  }
}
```

Không trả SQL error trực tiếp cho browser.

---

# 8. QUY TẮC JAVASCRIPT

Không tạo một file JS khổng lồ.

Tách:

```text
engine.js
ide.js
exam.js
course.js
teacher.js
```

Các module giao tiếp bằng object global có namespace.

Ví dụ:

```js
window.App = window.App || {};

App.API = {};
App.IDE = {};
App.Exam = {};
```

Không tạo quá nhiều global function.

---

# 9. QUY TẮC CSS

Không dùng framework.

Tạo biến CSS chung:

```css
:root {
    --bg-primary: #111;
    --bg-secondary: #1a1a1a;

    --text-primary: #fff;
    --text-secondary: #aaa;

    --space-xs: 4px;
    --space-sm: 8px;
    --space-md: 16px;
    --space-lg: 24px;

    --radius-sm: 4px;
    --radius-md: 8px;
}
```

Component class:

```text
.btn
.card
.badge
.table
.modal
.sidebar
.tabs
.form-control
```

Không copy CSS giữa các trang.

---

# 10. MVP NÊN DỪNG Ở ĐÂU

MVP không cần toàn bộ roadmap.

MVP nên hoàn thành:

```text
Phase 0
Phase 1
Phase 2
Phase 3
Phase 4
Phase 5
Phase 6
Phase 7
Phase 8
Phase 9
Phase 14
Phase 15
```

Sau đó hệ thống đã đủ để:

- login
- có student/teacher
- quản lý course
- quản lý class
- học lesson
- làm coding exercise
- lưu progress
- giao assignment
- xem kết quả
- teacher theo dõi sinh viên

Chưa cần:

```text
server judge production
exam proctoring mạnh
visualization
AI
```

---

# 11. THỨ TỰ COMMIT ĐỀ XUẤT

Mỗi commit chỉ nên có một mục tiêu rõ.

Ví dụ:

```text
refactor: split database schema and seed logic

feat: add schema migration system

feat: add users and session authentication

feat: add role middleware

feat: add course management

feat: add class and enrollment

feat: add chapter and lesson model

feat: migrate exercises to course structure

feat: add public and hidden test case schema

feat: add student exercise progress

feat: add attempt tracking

feat: add assignment management

feat: add student dashboard

feat: add teacher dashboard

feat: add exam schema

feat: add exam sessions

feat: add exam event monitoring
```

---

# 12. CHECKLIST CHO AGENT SAU MỖI PHASE

Sau mỗi phase Agent phải:

## 1. Build

```bash
go build ./...
```

## 2. Test

```bash
go test ./...
```

## 3. Run

```bash
go run ./cmd/web
```

## 4. Smoke Test

Kiểm tra route chính.

## 5. Database

Kiểm tra:

```text
migration applied
foreign key
data count
```

## 6. Regression

Đặc biệt phải kiểm tra lại:

```text
Run Python
Public Tests
Function whitelist
IDE
Function search
```

## 7. Update docs

Cập nhật:

```text
INDEX.md
README.md
```

nếu có thay đổi cấu trúc.

---

# 13. QUY TẮC KHÔNG ĐƯỢC VI PHẠM

Agent không được tự ý:

```text
thêm React
thêm Vue
thêm Tailwind
thêm Bootstrap
thêm Gin
thêm Echo
thêm Fiber
thêm ORM
```

Không thay:

```text
Skulpt
```

nếu chưa có lý do kỹ thuật rõ ràng.

Không viết lại project chỉ để "clean architecture".

Ưu tiên:

```text
incremental refactor
backward compatible
small commits
test after each phase
```

---

# 14. ƯU TIÊN KIẾN TRÚC

Thứ tự ưu tiên:

```text
1. Đúng
2. Đơn giản
3. Dễ bảo trì
4. Dễ mở rộng
5. Hiệu năng
```

Không over-engineer.

SQLite đủ cho giai đoạn đầu.

Go standard library đủ cho backend.

HTML server-side rendering đủ cho phần lớn UI.

JavaScript chỉ dùng khi cần interaction.

---

# 15. ĐỊNH HƯỚNG SAU MVP

Sau khi MVP ổn định mới xem xét:

```text
Algorithm visualization
Exam server judge
Sandbox
Learning analytics
Question randomization
Question pool
CSV import students
Export grades
Course cloning
Semester archive
AI tutor
AI feedback
```

AI không phải core system.

Core system phải hoạt động hoàn chỉnh mà không có AI.

---

# 16. KIẾN TRÚC CUỐI CÙNG

```text
                         Browser
                            |
           +----------------+----------------+
           |                                 |
        HTML/CSS                       Vanilla JS
           |                                 |
           |                        +--------+--------+
           |                        |                 |
           |                     Skulpt            Fetch API
           |                        |                 |
           +------------------------+-----------------+
                            |
                          HTTP
                            |
                            v
                       Go net/http
                            |
             +--------------+--------------+
             |              |              |
            Auth          Service        Handler
                            |
                            v
                       Repository
                            |
                            v
                      database/sql
                            |
                            v
                         SQLite
```

Practice execution:

```text
Student code
     |
     v
Skulpt
     |
     v
Public Tests
```

Official submission:

```text
Student code
     |
     v
Go Backend
     |
     v
Judge Interface
     |
     v
Hidden Tests
     |
     v
Submission Result
```

---

# 17. KẾT LUẬN TRIỂN KHAI

Không nên biến `web-python` thành một LMS tổng quát.

Mục tiêu nên là:

> **Programming Learning & Assessment Platform**

Hệ thống tập trung vào:

```text
Học khái niệm
      |
      v
Xem code
      |
      v
Chạy code trực tiếp
      |
      v
Luyện tập
      |
      v
Public Tests
      |
      v
Theo dõi tiến độ
      |
      v
Assignment
      |
      v
Exam
      |
      v
Assessment
```

Tài sản quan trọng nhất của source hiện tại là:

```text
IDE
Skulpt
Function whitelist
Exercise UI
Public test runner
```

Không nên viết lại các phần này nếu chưa cần.

Trọng tâm phát triển tiếp theo phải là:

```text
User
Course
Class
Lesson
Progress
Assignment
Submission
Exam
Teacher Analytics
```

Đây là phần biến project từ:

```text
Web Python IDE
```

thành:

```text
Nền tảng học và đánh giá lập trình hoàn chỉnh.
```
