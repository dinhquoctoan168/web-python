package database

import (
	"database/sql"
	"log"
)

type seedTestCase struct {
	CallExpression string
	ExpectedOutput string
	IsHidden       int
}

type seedExercise struct {
	Title             string
	Difficulty        string
	Description       string
	InitialCode       string
	SolutionCode      string
	SolutionHint      string
	VisualizationType string
	TestCases         []seedTestCase
}

type seedLesson struct {
	Title       string
	ContentHTML string
	Exercise    *seedExercise
}

type seedChapter struct {
	Title       string
	Description string
	Lessons     []seedLesson
}

// seedComprehensiveCourses nạp chương trình học và bài tập chi tiết cho cả 3 môn học (Phase 22)
func seedComprehensiveCourses(db *sql.DB) error {
	// 1. Lấy hoặc đảm bảo các môn học đã tồn tại
	courseMap := make(map[string]int)
	rows, err := db.Query("SELECT id, code FROM courses")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var code string
		if err := rows.Scan(&id, &code); err == nil {
			courseMap[code] = id
		}
	}

	// Đảm bảo đủ 3 môn
	if _, ok := courseMap["PY101"]; !ok {
		res, err := db.Exec(`INSERT INTO courses (code, name, description, status) VALUES ('PY101', 'Python cơ bản', 'Nhập môn lập trình Python', 'active')`)
		if err == nil {
			id, _ := res.LastInsertId()
			courseMap["PY101"] = int(id)
		}
	}
	if _, ok := courseMap["PRG201"]; !ok {
		res, err := db.Exec(`INSERT INTO courses (code, name, description, status) VALUES ('PRG201', 'Kỹ thuật lập trình', 'Kỹ thuật lập trình nâng cao', 'active')`)
		if err == nil {
			id, _ := res.LastInsertId()
			courseMap["PRG201"] = int(id)
		}
	}
	if _, ok := courseMap["DSA301"]; !ok {
		res, err := db.Exec(`INSERT INTO courses (code, name, description, status) VALUES ('DSA301', 'Cấu trúc dữ liệu và giải thuật', 'CTDL & Giải thuật thực chiến', 'active')`)
		if err == nil {
			id, _ := res.LastInsertId()
			courseMap["DSA301"] = int(id)
		}
	}

	// 2. Định nghĩa cấu trúc học liệu chuẩn hóa theo ROADMAP Phase 22
	// Môn 1: Python cơ bản (PY101)
	pyChapters := []seedChapter{
		{
			Title:       "Chương 1: Biến & Kiểu dữ liệu",
			Description: "Khái niệm biến, quy tắc đặt tên và các kiểu dữ liệu nền tảng",
			Lessons: []seedLesson{
				{
					Title: "1.1 Khai báo biến và Quy tắc đặt tên",
					ContentHTML: `<p>Trong Python, biến là định danh lưu trữ dữ liệu trong bộ nhớ mà không cần khai báo trước kiểu dữ liệu.</p>
<h4>Quy tắc:</h4>
<ul>
<li>Tên biến bắt đầu bằng ký tự hoặc dấu gạch dưới (_).</li>
<li>Phân biệt chữ hoa và chữ thường (case-sensitive).</li>
</ul>
<pre><code class="language-python">x = 10
greeting = "Xin chào"
_pi = 3.14159
print(greeting, x)</code></pre>`,
				},
				{
					Title: "1.2 Kiểu dữ liệu cơ bản (int, float, str, bool)",
					ContentHTML: `<p>Python cung cấp 4 kiểu dữ liệu nguyên thủy thường dùng nhất:</p>
<ul>
<li><strong>int</strong>: Số nguyên (ví dụ: 42, -7)</li>
<li><strong>float</strong>: Số thực dấu chấm động (ví dụ: 3.14, -0.5)</li>
<li><strong>str</strong>: Chuỗi ký tự (ví dụ: "Python")</li>
<li><strong>bool</strong>: Giá trị logic (True hoặc False)</li>
</ul>
<pre><code class="language-python">age = 20
score = 8.5
is_passed = score >= 5.0
print(type(age), type(score), type(is_passed))</code></pre>`,
				},
			},
		},
		{
			Title:       "Chương 2: Cấu trúc Rẽ nhánh (if)",
			Description: "Điều khiển luồng thực thi với câu lệnh điều kiện",
			Lessons: []seedLesson{
				{
					Title: "2.1 Câu lệnh rẽ nhánh if - elif - else",
					ContentHTML: `<p>Câu lệnh rẽ nhánh giúp chương trình đưa ra quyết định dựa trên điều kiện đúng/sai.</p>
<pre><code class="language-python">n = 7
if n % 2 == 0:
    print("Số chẵn")
else:
    print("Số lẻ")</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Kiểm tra số chẵn lẻ",
						Difficulty:   "Dễ",
						Description:  "Viết hàm `is_even(n)` nhận vào số nguyên `n`, trả về `True` nếu `n` là số chẵn, ngược lại trả về `False`.",
						InitialCode:  "def is_even(n):\n    # Viết code của bạn tại đây\n    pass\n",
						SolutionCode: "def is_even(n):\n    return n % 2 == 0\n",
						SolutionHint: "Sử dụng toán tử chia lấy dư `% 2` để kiểm tra số chẵn.",
						TestCases: []seedTestCase{
							{CallExpression: "is_even(4)", ExpectedOutput: "True", IsHidden: 0},
							{CallExpression: "is_even(7)", ExpectedOutput: "False", IsHidden: 0},
							{CallExpression: "is_even(0)", ExpectedOutput: "True", IsHidden: 1},
							{CallExpression: "is_even(-2)", ExpectedOutput: "True", IsHidden: 1},
						},
					},
				},
			},
		},
		{
			Title:       "Chương 3: Vòng lặp (for & while)",
			Description: "Lặp lại công việc với vòng lặp for và while",
			Lessons: []seedLesson{
				{
					Title: "3.1 Vòng lặp for và hàm range()",
					ContentHTML: `<p>Vòng lặp <code>for</code> thường được dùng khi biết trước số lần lặp hoặc duyệt qua một tập hợp.</p>
<pre><code class="language-python">total = 0
for i in range(1, 6):
    total += i
print("Tổng:", total)</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Tính tổng từ 1 đến N",
						Difficulty:   "Dễ",
						Description:  "Viết hàm `sum_to_n(n)` nhận vào số nguyên dương `n`, trả về tổng các số từ `1` đến `n`.",
						InitialCode:  "def sum_to_n(n):\n    # Viết code tại đây\n    pass\n",
						SolutionCode: "def sum_to_n(n):\n    return sum(range(1, n + 1))\n",
						SolutionHint: "Dùng vòng lặp for với range(1, n + 1) hoặc công thức n*(n+1)//2.",
						TestCases: []seedTestCase{
							{CallExpression: "sum_to_n(5)", ExpectedOutput: "15", IsHidden: 0},
							{CallExpression: "sum_to_n(10)", ExpectedOutput: "55", IsHidden: 0},
							{CallExpression: "sum_to_n(1)", ExpectedOutput: "1", IsHidden: 1},
							{CallExpression: "sum_to_n(100)", ExpectedOutput: "5050", IsHidden: 1},
						},
					},
				},
				{
					Title: "3.2 Vòng lặp while và điều kiện dừng",
					ContentHTML: `<p>Vòng lặp <code>while</code> thực thi khối lệnh chừng nào điều kiện logic còn mang giá trị <code>True</code>.</p>
<pre><code class="language-python">count = 5
while count > 0:
    print(count)
    count -= 1</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Đếm số chữ số nguyên dương",
						Difficulty:   "Dễ",
						Description:  "Viết hàm `count_digits(n)` nhận vào số nguyên dương `n`, trả về số chữ số của `n` sử dụng vòng lặp while.",
						InitialCode:  "def count_digits(n):\n    pass\n",
						SolutionCode: "def count_digits(n):\n    if n == 0:\n        return 1\n    count = 0\n    while n > 0:\n        count += 1\n        n //= 10\n    return count\n",
						SolutionHint: "Chia nguyên `n //= 10` trong mỗi lần lặp cho tới khi `n == 0`.",
						TestCases: []seedTestCase{
							{CallExpression: "count_digits(12345)", ExpectedOutput: "5", IsHidden: 0},
							{CallExpression: "count_digits(7)", ExpectedOutput: "1", IsHidden: 0},
							{CallExpression: "count_digits(100000)", ExpectedOutput: "6", IsHidden: 1},
						},
					},
				},
			},
		},
		{
			Title:       "Chương 4: Hàm (Functions)",
			Description: "Định nghĩa, tham số và giá trị trả về của hàm",
			Lessons: []seedLesson{
				{
					Title: "4.1 Định nghĩa và gọi hàm (def & return)",
					ContentHTML: `<p>Hàm giúp đóng gói mã nguồn thành các khối logic có thể tái sử dụng.</p>
<pre><code class="language-python">def add(a, b):
    return a + b

res = add(3, 4)
print("Kết quả:", res)</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Tính diện tích hình chữ nhật",
						Difficulty:   "Dễ",
						Description:  "Viết hàm `rectangle_area(width, height)` trả về diện tích hình chữ nhật.",
						InitialCode:  "def rectangle_area(width, height):\n    pass\n",
						SolutionCode: "def rectangle_area(width, height):\n    return width * height\n",
						SolutionHint: "Diện tích = chiều rộng * chiều cao.",
						TestCases: []seedTestCase{
							{CallExpression: "rectangle_area(5, 4)", ExpectedOutput: "20", IsHidden: 0},
							{CallExpression: "rectangle_area(3.5, 2)", ExpectedOutput: "7.0", IsHidden: 0},
							{CallExpression: "rectangle_area(10, 10)", ExpectedOutput: "100", IsHidden: 1},
						},
					},
				},
			},
		},
		{
			Title:       "Chương 5: Danh sách (List) & Từ điển (Dictionary)",
			Description: "Các cấu trúc dữ liệu cơ bản dạng mảng và bảng băm ánh xạ",
			Lessons: []seedLesson{
				{
					Title: "5.1 Thao tác với List",
					ContentHTML: `<p>List là danh sách phần tử có thứ tự và có thể thay đổi (mutable).</p>
<pre><code class="language-python">arr = [1, 2, 3]
arr.append(4)
print(arr[0], len(arr))</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Tìm phần tử lớn nhất trong List",
						Difficulty:   "Dễ",
						Description:  "Viết hàm `find_max_element(numbers)` nhận vào một danh sách các số không rỗng và trả về giá trị lớn nhất.",
						InitialCode:  "def find_max_element(numbers):\n    pass\n",
						SolutionCode: "def find_max_element(numbers):\n    return max(numbers)\n",
						SolutionHint: "Dùng hàm built-in max() hoặc duyệt vòng lặp gán biến max_val.",
						TestCases: []seedTestCase{
							{CallExpression: "find_max_element([1, 5, 2, 9, 3])", ExpectedOutput: "9", IsHidden: 0},
							{CallExpression: "find_max_element([-10, -5, -20])", ExpectedOutput: "-5", IsHidden: 0},
							{CallExpression: "find_max_element([42])", ExpectedOutput: "42", IsHidden: 1},
						},
					},
				},
				{
					Title: "5.2 Làm việc với Dictionary (Key-Value)",
					ContentHTML: `<p>Dictionary lưu trữ dữ liệu theo cặp khóa - giá trị (Key - Value) với tốc độ tra cứu O(1).</p>
<pre><code class="language-python">student = {"name": "An", "score": 9.5}
print(student["name"])</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Đếm tần suất ký tự",
						Difficulty:   "Trung bình",
						Description:  "Viết hàm `char_frequency(s)` nhận vào chuỗi `s` và trả về một dictionary lưu số lần xuất hiện của từng ký tự.",
						InitialCode:  "def char_frequency(s):\n    pass\n",
						SolutionCode: "def char_frequency(s):\n    d = {}\n    for c in s:\n        d[c] = d.get(c, 0) + 1\n    return d\n",
						SolutionHint: "Dùng dictionary với phương thức `d.get(c, 0) + 1`.",
						TestCases: []seedTestCase{
							{CallExpression: "char_frequency('aba')", ExpectedOutput: "{'a': 2, 'b': 1}", IsHidden: 0},
							{CallExpression: "char_frequency('hello')", ExpectedOutput: "{'h': 1, 'e': 1, 'l': 2, 'o': 1}", IsHidden: 0},
						},
					},
				},
			},
		},
		{
			Title:       "Chương 6: Thao tác Tệp (File) & Xử lý Ngoại lệ (Exception)",
			Description: "Đọc ghi file an toàn và bắt lỗi bằng khối lệnh try - except",
			Lessons: []seedLesson{
				{
					Title: "6.1 Đọc và ghi tệp trong Python",
					ContentHTML: `<p>Sử dụng ngữ cảnh <code>with open(...)</code> để tự động đóng tệp an toàn.</p>
<pre><code class="language-python"># Mẫu ghi tệp
with open("data.txt", "w", encoding="utf-8") as f:
    f.write("Xin chào Python!")</code></pre>`,
				},
				{
					Title: "6.2 Bắt và xử lý ngoại lệ (try - except)",
					ContentHTML: `<p>Khối lệnh <code>try ... except</code> ngăn ngừa ứng dụng bị crash khi gặp lỗi bất ngờ lúc thực thi.</p>
<pre><code class="language-python">try:
    val = int("abc")
except ValueError as e:
    print("Lỗi chuyển đổi kiểu dữ liệu:", e)</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Phép chia an toàn bắt lỗi ZeroDivisionError",
						Difficulty:   "Dễ",
						Description:  "Viết hàm `safe_division(a, b)` thực hiện phép chia `a / b`. Nếu `b == 0`, bắt lỗi `ZeroDivisionError` và trả về `None`.",
						InitialCode:  "def safe_division(a, b):\n    pass\n",
						SolutionCode: "def safe_division(a, b):\n    try:\n        return a / b\n    except ZeroDivisionError:\n        return None\n",
						SolutionHint: "Dùng khối try ... except ZeroDivisionError.",
						TestCases: []seedTestCase{
							{CallExpression: "safe_division(10, 2)", ExpectedOutput: "5.0", IsHidden: 0},
							{CallExpression: "safe_division(5, 0)", ExpectedOutput: "None", IsHidden: 0},
							{CallExpression: "safe_division(0, 5)", ExpectedOutput: "0.0", IsHidden: 1},
						},
					},
				},
			},
		},
	}

	// Môn 2: Kỹ thuật lập trình (PRG201)
	prgChapters := []seedChapter{
		{
			Title:       "Chương 1: Thiết kế Hàm & Module",
			Description: "Kỹ thuật viết hàm chuẩn mực và tổ chức dự án",
			Lessons: []seedLesson{
				{
					Title: "1.1 Kỹ thuật thiết kế hàm (Function Design & Type Hints)",
					ContentHTML: `<p>Thiết kế hàm chuyên nghiệp tuân theo nguyên lý Single Responsibility (Đơn nhiệm), sử dụng Type Hints và Docstrings rõ ràng.</p>
<pre><code class="language-python">def calculate_discount(price: float, rate: float) -> float:
    """Tính giá tiền sau khi giảm giá."""
    if not (0 <= rate <= 1):
        raise ValueError("Tỷ lệ giảm giá phải từ 0 đến 1")
    return price * (1 - rate)</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Chuẩn hóa họ tên",
						Difficulty:   "Dễ",
						Description:  "Viết hàm `normalize_full_name(first_name: str, last_name: str) -> str` trả về họ và tên đầy đủ viết hoa chữ cái đầu mỗi từ.",
						InitialCode:  "def normalize_full_name(first_name: str, last_name: str) -> str:\n    pass\n",
						SolutionCode: "def normalize_full_name(first_name: str, last_name: str) -> str:\n    return f\"{first_name.strip().title()} {last_name.strip().title()}\"\n",
						SolutionHint: "Sử dụng các phương thức chuỗi strip() và title().",
						TestCases: []seedTestCase{
							{CallExpression: "normalize_full_name('nguyen', 'van an')", ExpectedOutput: "'Nguyen Van An'", IsHidden: 0},
							{CallExpression: "normalize_full_name(' tran ', 'ba ')", ExpectedOutput: "'Tran Ba'", IsHidden: 1},
						},
					},
				},
				{
					Title: "1.2 Tổ chức Module và Quản lý Thư viện",
					ContentHTML: `<p>Mỗi tệp <code>.py</code> là một module độc lập. Sử dụng <code>import</code> để tái sử dụng mã nguồn và tránh xung đột namespace.</p>
<pre><code class="language-python">import math
print("Căn bậc hai của 16:", math.sqrt(16))</code></pre>`,
				},
			},
		},
		{
			Title:       "Chương 2: Xử lý Lỗi & Ngoại lệ Chuyên sâu (Error Handling)",
			Description: "Thiết kế hệ thống chịu lỗi và Custom Exceptions",
			Lessons: []seedLesson{
				{
					Title: "2.1 Ngoại lệ tùy biến (Custom Exceptions)",
					ContentHTML: `<p>Tạo các lớp ngoại lệ kế thừa từ <code>Exception</code> để biểu diễn lỗi nghiệp vụ rõ nghĩa.</p>
<pre><code class="language-python">class ValidationError(Exception):
    pass

def set_age(age: int):
    if age < 0:
        raise ValidationError("Tuổi không thể là số âm")</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Kiểm tra mật khẩu hợp lệ",
						Difficulty:   "Trung bình",
						Description:  "Viết hàm `is_strong_password(pwd: str) -> bool` kiểm tra mật khẩu có độ dài ít nhất 8 ký tự, có ít nhất một chữ hoa và một số hay không.",
						InitialCode:  "def is_strong_password(pwd: str) -> bool:\n    pass\n",
						SolutionCode: "def is_strong_password(pwd: str) -> bool:\n    if len(pwd) < 8:\n        return False\n    has_upper = any(c.isupper() for c in pwd)\n    has_digit = any(c.isdigit() for c in pwd)\n    return has_upper and has_digit\n",
						SolutionHint: "Dùng các hàm c.isupper(), c.isdigit() kết hợp hàm any().",
						TestCases: []seedTestCase{
							{CallExpression: "is_strong_password('Pass1234')", ExpectedOutput: "True", IsHidden: 0},
							{CallExpression: "is_strong_password('weak')", ExpectedOutput: "False", IsHidden: 0},
							{CallExpression: "is_strong_password('NoDigitsHere')", ExpectedOutput: "False", IsHidden: 1},
							{CallExpression: "is_strong_password('12345678')", ExpectedOutput: "False", IsHidden: 1},
						},
					},
				},
			},
		},
		{
			Title:       "Chương 3: Lập trình Hướng đối tượng (OOP)",
			Description: "Class, Object, Constructor, Kế thừa và Đa hình",
			Lessons: []seedLesson{
				{
					Title: "3.1 Xây dựng Lớp và Đối tượng",
					ContentHTML: `<p>OOP giúp mô hình hóa thực thể trong thế giới thực thành các Class gồm Thuộc tính và Phương thức.</p>
<pre><code class="language-python">class Person:
    def __init__(self, name: str, age: int):
        self.name = name
        self.age = age

    def introduce(self) -> str:
        return f"Tôi là {self.name}, {self.age} tuổi."

p = Person("Bình", 21)
print(p.introduce())</code></pre>`,
					Exercise: &seedExercise{
						Title:        "Xây dựng lớp Circle",
						Difficulty:   "Trung bình",
						Description:  "Viết hàm `create_circle_info(radius: float)` tạo ra lớp `Circle` có phương thức `area()` tính diện tích hình tròn (lấy số pi = 3.14). Hàm trả về diện tích làm tròn 2 chữ số.",
						InitialCode:  "def create_circle_info(radius: float) -> float:\n    # Viết code tạo class hoặc tính toán tại đây\n    pass\n",
						SolutionCode: "def create_circle_info(radius: float) -> float:\n    class Circle:\n        def __init__(self, r):\n            self.r = r\n        def area(self):\n            return round(3.14 * self.r * self.r, 2)\n    c = Circle(radius)\n    return c.area()\n",
						SolutionHint: "Công thức diện tích S = 3.14 * r * r.",
						TestCases: []seedTestCase{
							{CallExpression: "create_circle_info(2.0)", ExpectedOutput: "12.56", IsHidden: 0},
							{CallExpression: "create_circle_info(5.0)", ExpectedOutput: "78.5", IsHidden: 0},
							{CallExpression: "create_circle_info(1.0)", ExpectedOutput: "3.14", IsHidden: 1},
						},
					},
				},
			},
		},
		{
			Title:       "Chương 4: Kiểm thử & Gỡ lỗi (Testing & Debugging)",
			Description: "Viết test case tự động và kỹ thuật debug hệ thống",
			Lessons: []seedLesson{
				{
					Title: "4.1 Kiểm thử phần mềm với Unit Testing",
					ContentHTML: `<p>Kiểm thử đơn vị (Unit Testing) kiểm chứng từng hàm nhỏ trong chương trình hoạt động đúng theo đặc tả kỹ thuật.</p>
<pre><code class="language-python">def is_prime(n: int) -> bool:
    if n < 2:
        return False
    for i in range(2, int(n ** 0.5) + 1):
        if n % i == 0:
            return False
    return True

# Test assertions
assert is_prime(2) == True
assert is_prime(4) == False
print("Tất cả test case kiểm thử đều thành công!")</code></pre>`,
				},
				{
					Title: "4.2 Kỹ thuật Gỡ lỗi (Debugging)",
					ContentHTML: `<p>Gỡ lỗi là kỹ năng cô lập nguyên nhân sinh ra bug. Áp dụng kỹ thuật in vết (tracing), logging và breakpoint.</p>
<pre><code class="language-python">import logging
logging.basicConfig(level=logging.INFO)
logging.info("Hệ thống khởi động an toàn")</code></pre>`,
				},
			},
		},
	}

	// Môn 3: Cấu trúc dữ liệu và giải thuật (DSA301)
	dsaChapters := []seedChapter{
		{
			Title:       "Chương 1: Cấu trúc Dữ liệu Tuyến tính (Linear DS)",
			Description: "Mảng, Ngăn xếp, Hàng đợi và Danh sách liên kết",
			Lessons: []seedLesson{
				{
					Title: "1.1 Mảng (Array) và Mảng động",
					ContentHTML: `<p>Mảng lưu trữ các phần tử liên tiếp trong bộ nhớ, cho phép truy xuất ngẫu nhiên O(1) theo chỉ số.</p>`,
					Exercise: &seedExercise{
						Title:             "Đảo ngược danh sách (Reverse Array)",
						Difficulty:        "Dễ",
						Description:       "Viết hàm `reverse_list(arr)` nhận vào danh sách và trả về danh sách có thứ tự ngược lại.",
						InitialCode:       "def reverse_list(arr):\n    pass\n",
						SolutionCode:      "def reverse_list(arr):\n    return arr[::-1]\n",
						SolutionHint:      "Dùng slicing `arr[::-1]` hoặc hai con trỏ swap.",
						VisualizationType: "array",
						TestCases: []seedTestCase{
							{CallExpression: "reverse_list([1, 2, 3, 4])", ExpectedOutput: "[4, 3, 2, 1]", IsHidden: 0},
							{CallExpression: "reverse_list(['a', 'b'])", ExpectedOutput: "['b', 'a']", IsHidden: 0},
							{CallExpression: "reverse_list([])", ExpectedOutput: "[]", IsHidden: 1},
						},
					},
				},
				{
					Title: "1.2 Ngăn xếp (Stack - LIFO)",
					ContentHTML: `<p>Ngăn xếp hoạt động theo cơ chế Vào sau Ra trước (Last In, First Out). Phù hợp giải bài toán cặp ngoặc, khử đệ quy.</p>`,
					Exercise: &seedExercise{
						Title:             "Kiểm tra cặp ngoặc hợp lệ bằng Stack",
						Difficulty:        "Trung bình",
						Description:       "Viết hàm `is_valid_brackets(s: str) -> bool` kiểm tra chuỗi chứa các dấu ngoặc `()`, `[]`, `{}` có đóng mở hợp lệ hay không.",
						InitialCode:       "def is_valid_brackets(s: str) -> bool:\n    pass\n",
						SolutionCode:      "def is_valid_brackets(s: str) -> bool:\n    stack = []\n    mapping = {')': '(', ']': '[', '}': '{'}\n    for char in s:\n        if char in mapping.values():\n            stack.append(char)\n        elif char in mapping:\n            if not stack or stack.pop() != mapping[char]:\n                return False\n    return len(stack) == 0\n",
						SolutionHint:      "Duyệt từng ký tự, gặp mở ngoặc thì push vào stack, gặp đóng ngoặc thì pop và so sánh.",
						VisualizationType: "stack",
						TestCases: []seedTestCase{
							{CallExpression: "is_valid_brackets('()[]{}')", ExpectedOutput: "True", IsHidden: 0},
							{CallExpression: "is_valid_brackets('(]')", ExpectedOutput: "False", IsHidden: 0},
							{CallExpression: "is_valid_brackets('([{}])')", ExpectedOutput: "True", IsHidden: 1},
							{CallExpression: "is_valid_brackets('(((')", ExpectedOutput: "False", IsHidden: 1},
						},
					},
				},
				{
					Title: "1.3 Hàng đợi (Queue - FIFO)",
					ContentHTML: `<p>Hàng đợi hoạt động theo nguyên lý Vào trước Ra trước (First In, First Out), là nền tảng của duyệt cây/đồ thị BFS.</p>`,
				},
				{
					Title: "1.4 Danh sách liên kết (Linked List)",
					ContentHTML: `<p>Mỗi nút (Node) lưu dữ liệu và con trỏ trỏ tới phần tử tiếp theo, cho phép chèn/xóa O(1) ở đầu danh sách.</p>`,
				},
			},
		},
		{
			Title:       "Chương 2: Cấu trúc Dữ liệu Phi tuyến (Trees & Graphs)",
			Description: "Cây nhị phân tìm kiếm và Đồ thị",
			Lessons: []seedLesson{
				{
					Title: "2.1 Cây nhị phân và BST (Binary Search Tree)",
					ContentHTML: `<p>Cây nhị phân tìm kiếm đảm bảo mọi nút con bên trái có giá trị nhỏ hơn nút gốc, và nút con bên phải có giá trị lớn hơn.</p>`,
				},
				{
					Title: "2.2 Đồ thị và Thuật toán duyệt BFS / DFS",
					ContentHTML: `<p>Đồ thị biểu diễn mối quan hệ giữa các đỉnh (Vertices) và cạnh (Edges). BFS duyệt theo chiều rộng, DFS duyệt theo chiều sâu.</p>`,
				},
			},
		},
		{
			Title:       "Chương 3: Thuật toán Sắp xếp & Tìm kiếm",
			Description: "Sorting & Searching Algorithms",
			Lessons: []seedLesson{
				{
					Title: "3.1 Các giải thuật Sắp xếp (Sorting)",
					ContentHTML: `<p>Sắp xếp là bài toán cơ bản nhất trong khoa học máy tính: Bubble Sort O(N^2), Quick Sort O(N log N), Merge Sort O(N log N).</p>`,
					Exercise: &seedExercise{
						Title:             "Cài đặt thuật toán Bubble Sort",
						Difficulty:        "Trung bình",
						Description:       "Viết hàm `bubble_sort(arr)` thực hiện sắp xếp danh sách các số nguyên theo thứ tự tăng dần.",
						InitialCode:       "def bubble_sort(arr):\n    pass\n",
						SolutionCode:      "def bubble_sort(arr):\n    res = list(arr)\n    n = len(res)\n    for i in range(n):\n        for j in range(0, n - i - 1):\n            if res[j] > res[j + 1]:\n                res[j], res[j + 1] = res[j + 1], res[j]\n    return res\n",
						SolutionHint:      "So sánh từng cặp phần tử liền kề và hoán đổi nếu sai thứ tự.",
						VisualizationType: "sorting",
						TestCases: []seedTestCase{
							{CallExpression: "bubble_sort([5, 2, 9, 1, 3])", ExpectedOutput: "[1, 2, 3, 5, 9]", IsHidden: 0},
							{CallExpression: "bubble_sort([3, 3, 1])", ExpectedOutput: "[1, 3, 3]", IsHidden: 0},
							{CallExpression: "bubble_sort([])", ExpectedOutput: "[]", IsHidden: 1},
						},
					},
				},
				{
					Title: "3.2 Tìm kiếm nhị phân (Binary Search)",
					ContentHTML: `<p>Tìm kiếm nhị phân chia đôi không gian tìm kiếm ở mỗi bước, đạt độ phức tạp O(log N) trên mảng đã sắp xếp.</p>`,
					Exercise: &seedExercise{
						Title:             "Thuật toán Tìm kiếm nhị phân",
						Difficulty:        "Trung bình",
						Description:       "Viết hàm `binary_search(arr, target)` nhận mảng đã sắp xếp tăng dần và giá trị cần tìm. Trả về chỉ số (0-based) nếu thấy, ngược lại trả về -1.",
						InitialCode:       "def binary_search(arr, target):\n    pass\n",
						SolutionCode:      "def binary_search(arr, target):\n    left, right = 0, len(arr) - 1\n    while left <= right:\n        mid = (left + right) // 2\n        if arr[mid] == target:\n            return mid\n        elif arr[mid] < target:\n            left = mid + 1\n        else:\n            right = mid - 1\n    return -1\n",
						SolutionHint:      "Duyệt với left, right và chia đôi mid = (left + right) // 2.",
						VisualizationType: "binary_search",
						TestCases: []seedTestCase{
							{CallExpression: "binary_search([2, 5, 8, 12, 16], 8)", ExpectedOutput: "2", IsHidden: 0},
							{CallExpression: "binary_search([1, 3, 5], 4)", ExpectedOutput: "-1", IsHidden: 0},
							{CallExpression: "binary_search([10], 10)", ExpectedOutput: "0", IsHidden: 1},
						},
					},
				},
			},
		},
		{
			Title:       "Chương 4: Kỹ thuật Thiết kế Giải thuật",
			Description: "Đệ quy, Tham lam và Quy hoạch động",
			Lessons: []seedLesson{
				{
					Title: "4.1 Thuật toán Đệ quy (Recursion)",
					ContentHTML: `<p>Đệ quy là phương pháp giải quyết bài toán bằng cách gọi lại chính nó với bài toán con nhỏ hơn kèm theo trường hợp cơ sở (base case).</p>`,
					Exercise: &seedExercise{
						Title:        "Tính giai thừa đệ quy",
						Difficulty:   "Dễ",
						Description:  "Viết hàm đệ quy `factorial(n: int) -> int` tính `n!`. Quy ước `0! = 1`.",
						InitialCode:  "def factorial(n: int) -> int:\n    pass\n",
						SolutionCode: "def factorial(n: int) -> int:\n    if n <= 1:\n        return 1\n    return n * factorial(n - 1)\n",
						SolutionHint: "Base case: if n <= 1 return 1. Recursive case: n * factorial(n - 1).",
						TestCases: []seedTestCase{
							{CallExpression: "factorial(5)", ExpectedOutput: "120", IsHidden: 0},
							{CallExpression: "factorial(0)", ExpectedOutput: "1", IsHidden: 0},
							{CallExpression: "factorial(6)", ExpectedOutput: "720", IsHidden: 1},
						},
					},
				},
				{
					Title: "4.2 Giải thuật Tham lam (Greedy)",
					ContentHTML: `<p>Thuật toán tham lam lựa chọn phương án tối ưu cục bộ tại mỗi bước với kỳ vọng đạt tối ưu toàn cục.</p>`,
				},
				{
					Title: "4.3 Quy hoạch động (Dynamic Programming)",
					ContentHTML: `<p>Quy hoạch động lưu lại kết quả của các bài toán con bị gối nhau (Overlapping Subproblems) để tránh tính toán trùng lặp.</p>`,
					Exercise: &seedExercise{
						Title:        "Số Fibonacci với Quy hoạch động",
						Difficulty:   "Trung bình",
						Description:  "Viết hàm `fibonacci_dp(n: int) -> int` trả về số Fibonacci thứ n (F(0)=0, F(1)=1, F(2)=1, F(3)=2...).",
						InitialCode:  "def fibonacci_dp(n: int) -> int:\n    pass\n",
						SolutionCode: "def fibonacci_dp(n: int) -> int:\n    if n <= 0:\n        return 0\n    if n == 1:\n        return 1\n    a, b = 0, 1\n    for _ in range(2, n + 1):\n        a, b = b, a + b\n    return b\n",
						SolutionHint: "Dùng 2 biến a, b lặp từ 2 tới n để đạt O(N) thời gian và O(1) không gian bộ nhớ.",
						TestCases: []seedTestCase{
							{CallExpression: "fibonacci_dp(7)", ExpectedOutput: "13", IsHidden: 0},
							{CallExpression: "fibonacci_dp(0)", ExpectedOutput: "0", IsHidden: 0},
							{CallExpression: "fibonacci_dp(10)", ExpectedOutput: "55", IsHidden: 1},
							{CallExpression: "fibonacci_dp(20)", ExpectedOutput: "6765", IsHidden: 1},
						},
					},
				},
			},
		},
	}

	// 3. Tiến hành chèn dữ liệu vào CSDL bảo đảm idempotency
	allCourseData := map[string][]seedChapter{
		"PY101":  pyChapters,
		"PRG201": prgChapters,
		"DSA301": dsaChapters,
	}

	for code, chapters := range allCourseData {
		courseID := courseMap[code]
		for cIdx, chap := range chapters {
			chapOrder := cIdx + 1
			var chapID int
			err := db.QueryRow(`SELECT id FROM chapters WHERE course_id = ? AND title = ?`, courseID, chap.Title).Scan(&chapID)
			if err == sql.ErrNoRows {
				res, err := db.Exec(`INSERT INTO chapters (course_id, title, description, order_num) VALUES (?, ?, ?, ?)`,
					courseID, chap.Title, chap.Description, chapOrder)
				if err != nil {
					return err
				}
				id, _ := res.LastInsertId()
				chapID = int(id)
			} else if err != nil {
				return err
			}

			for lIdx, les := range chap.Lessons {
				lesOrder := lIdx + 1
				var lesID int
				err := db.QueryRow(`SELECT id FROM lessons WHERE chapter_id = ? AND title = ?`, chapID, les.Title).Scan(&lesID)
				if err == sql.ErrNoRows {
					res, err := db.Exec(`INSERT INTO lessons (chapter_id, title, content_html, order_num, is_published) VALUES (?, ?, ?, ?, 1)`,
						chapID, les.Title, les.ContentHTML, lesOrder)
					if err != nil {
						return err
					}
					id, _ := res.LastInsertId()
					lesID = int(id)
				} else if err != nil {
					return err
				}

				// Đính kèm bài tập thực hành nếu có
				if les.Exercise != nil {
					var exID int
					err := db.QueryRow(`SELECT id FROM exercises WHERE course_id = ? AND title = ?`, courseID, les.Exercise.Title).Scan(&exID)
					if err == sql.ErrNoRows {
						visType := les.Exercise.VisualizationType
						if visType == "" {
							visType = "none"
						}
						res, err := db.Exec(`INSERT INTO exercises 
							(course_id, lesson_id, title, difficulty, description, initial_code, solution_code, solution_hint, visualization_type, status)
							VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'active')`,
							courseID, lesID, les.Exercise.Title, les.Exercise.Difficulty, les.Exercise.Description,
							les.Exercise.InitialCode, les.Exercise.SolutionCode, les.Exercise.SolutionHint, visType)
						if err != nil {
							return err
						}
						id, _ := res.LastInsertId()
						exID = int(id)

						for tcIdx, tc := range les.Exercise.TestCases {
							_, err := db.Exec(`INSERT INTO exercise_test_cases 
								(exercise_id, call_expression, expected_output, is_hidden, weight, order_num)
								VALUES (?, ?, ?, ?, 1.0, ?)`,
								exID, tc.CallExpression, tc.ExpectedOutput, tc.IsHidden, tcIdx+1)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	log.Println("Đã nạp thành công bộ dữ liệu học tập chi tiết cho 3 môn học (Phase 22).")
	return nil
}
