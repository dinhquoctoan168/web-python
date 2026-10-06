package database

import (
	"database/sql"
	"log"
)

func seedInitialData(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM topics").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // Đã có dữ liệu
	}

	log.Println("Đang nạp dữ liệu mẫu ban đầu cho môn Cơ sở dữ liệu và Giải thuật...")

	// 1. Thêm chủ đề
	topics := []struct {
		name, desc, icon string
		order            int
	}{
		{"Mảng & Danh sách (Array/List)", "Các thao tác cơ bản trên mảng, duyệt phần tử và tìm kiếm", "list", 1},
		{"Ngăn xếp & Hàng đợi (Stack & Queue)", "Cấu trúc LIFO và FIFO ứng dụng trong giải thuật", "layers", 2},
		{"Bảng băm & Bản ghi (Hash Table / Dict)", "Mô phỏng lưu trữ khóa - giá trị và bảng dữ liệu quan hệ", "database", 3},
		{"Cây nhị phân & Tìm kiếm (BST)", "Duyệt cây, tìm kiếm nhị phân và đệ quy", "git-branch", 4},
		{"Thuật toán sắp xếp (Sorting)", "Sắp xếp nổi bọt, chọn, chèn và sắp xếp nhanh", "bar-chart-2", 5},
	}

	for _, t := range topics {
		_, err := db.Exec("INSERT INTO topics (name, description, icon, order_num) VALUES (?, ?, ?, ?)",
			t.name, t.desc, t.icon, t.order)
		if err != nil {
			return err
		}
	}

	// 2. Thêm danh mục hàm tra cứu
	if err := seedFunctions(db); err != nil {
		return err
	}

	// 3. Thêm bài tập mẫu môn CSDL & Giải thuật
	exercises := []struct {
		topicID          int
		title            string
		difficulty       string
		desc             string
		initialCode      string
		allowedFunctions string
		testCases        string
		hint             string
	}{
		{
			topicID:    1,
			title:      "Tìm phần tử lớn nhất và nhỏ nhất trong Mảng",
			difficulty: "Dễ",
			desc: `Cho một danh sách các số nguyên. Hãy viết hàm find_min_max(arr) trả về tuple (min_val, max_val).
Yêu cầu: Không sử dụng hàm có sẵn min() hoặc max(), hãy tự duyệt mảng bằng vòng lặp for và so sánh.`,
			initialCode: `def find_min_max(arr):
    # TODO: Cài đặt giải thuật duyệt mảng tìm min và max tại đây
    # Trả về tuple (min_val, max_val), nếu rỗng trả về (None, None)
    pass

# Kiểm tra kết quả
test_arr = [12, 5, 7, 25, 3, 19]
print("Mảng kiểm tra:", test_arr)
print("Kết quả:", find_min_max(test_arr))
`,
			allowedFunctions: `["len", "range", "print"]`,
			testCases:        `[{"call": "find_min_max([12, 5, 7, 25, 3, 19])", "input": "[12, 5, 7, 25, 3, 19]", "expected": "(3, 25)"}, {"call": "find_min_max([-5, -1, -10])", "input": "[-5, -1, -10]", "expected": "(-10, -1)"}]`,
			hint:             "Khởi tạo min_val và max_val bằng phần tử đầu tiên arr[0], sau đó duyệt từ vị trí 1 đến hết mảng.",
		},
		{
			topicID:    2,
			title:      "Mô phỏng Ngăn xếp (Stack) kiểm tra đóng mở ngoặc",
			difficulty: "Trung bình",
			desc: `Ứng dụng cấu trúc Stack: Viết hàm is_valid_parentheses(s) kiểm tra chuỗi ngoặc gồm '(', ')', '{', '}', '[', ']' có hợp lệ hay không.
Chuỗi hợp lệ khi các dấu mở phải được đóng bởi cùng loại và theo đúng thứ tự LIFO.`,
			initialCode: `def is_valid_parentheses(s):
    # TODO: Mô phỏng Stack bằng list để kiểm tra tính hợp lệ của dấu ngoặc
    # Trả về True nếu hợp lệ, ngược lại False
    pass

# Kiểm tra
print("()[]{} ->", is_valid_parentheses("()[]{}"))
print("([)] ->", is_valid_parentheses("([)]"))
`,
			allowedFunctions: `["len", "append", "pop", "print"]`,
			testCases:        `[{"call": "is_valid_parentheses(\"()[]{}\")", "input": "\"()[]{}\"", "expected": "True"}, {"call": "is_valid_parentheses(\"([)]\")", "input": "\"([)]\"", "expected": "False"}]`,
			hint:             "Đẩy dấu mở vào Stack bằng append(). Khi gặp dấu đóng, dùng pop() để lấy phần tử đỉnh đối chiếu.",
		},
		{
			topicID:    3,
			title:      "Mô phỏng Bảng CSDL: Truy vấn SELECT và WHERE",
			difficulty: "Trung bình",
			desc: `Cho một danh sách các bản ghi (list of dict) đại diện cho bảng sinh viên trong CSDL:
sinh_vien = [
    {"id": 1, "name": "Nam", "gpa": 3.6, "dept": "CNTT"},
    {"id": 2, "name": "Hoa", "gpa": 3.1, "dept": "HTTT"},
    {"id": 3, "name": "Bình", "gpa": 3.8, "dept": "CNTT"}
]
Viết hàm query_students(students, min_gpa, dept) trả về danh sách tên sinh viên thỏa mãn cả 2 điều kiện gpa >= min_gpa và dept == dept.`,
			initialCode: `def query_students(students, min_gpa, dept):
    # TODO: Lọc bản ghi tương đương câu truy vấn:
    # SELECT name FROM students WHERE gpa >= min_gpa AND dept = dept
    pass

# Dữ liệu thử nghiệm
students = [
    {"id": 1, "name": "Nam", "gpa": 3.6, "dept": "CNTT"},
    {"id": 2, "name": "Hoa", "gpa": 3.1, "dept": "HTTT"},
    {"id": 3, "name": "Bình", "gpa": 3.8, "dept": "CNTT"},
    {"id": 4, "name": "Lan", "gpa": 2.8, "dept": "CNTT"}
]

print("Sinh viên CNTT có GPA >= 3.5:")
print(query_students(students, 3.5, "CNTT"))
`,
			allowedFunctions: `["len", "append", "print", "dict.get"]`,
			testCases:        `[{"call": "query_students(students, 3.5, 'CNTT')", "input": "students, min_gpa=3.5, dept='CNTT'", "expected": "['Nam', 'Bình']"}]`,
			hint:             "Duyệt qua từng dict trong danh sách, sử dụng s.get('key') để kiểm tra điều kiện WHERE.",
		},
		{
			topicID:    4,
			title:      "Tìm kiếm nhị phân (Binary Search)",
			difficulty: "Dễ",
			desc: `Cho một mảng số nguyên đã được sắp xếp tăng dần và một số target.
Viết hàm binary_search(arr, target) trả về chỉ số của target trong mảng. Nếu không tìm thấy, trả về -1. Độ phức tạp thời gian yêu cầu O(log N).`,
			initialCode: `def binary_search(arr, target):
    # TODO: Cài đặt giải thuật tìm kiếm nhị phân (Binary Search) O(log N)
    # Trả về chỉ số vị trí tìm thấy, hoặc -1 nếu không tìm thấy
    pass

# Kiểm tra
arr = [2, 5, 8, 12, 16, 23, 38, 56, 72, 91]
target = 23
print("Mảng:", arr)
print("Vị trí của", target, "là:", binary_search(arr, target))
`,
			allowedFunctions: `["len", "range", "print"]`,
			testCases:        `[{"call": "binary_search([2, 5, 8, 12, 16, 23, 38, 56, 72, 91], 23)", "input": "arr=[2, 5, 8, ...], target=23", "expected": "5"}, {"call": "binary_search([2, 5, 8, 12], 9)", "input": "arr=[2, 5, 8, 12], target=9", "expected": "-1"}]`,
			hint:             "Tính chỉ số giữa mid = (left + right) // 2, thu hẹp khoảng tìm kiếm sau mỗi bước so sánh.",
		},
	}

	for _, ex := range exercises {
		_, err := db.Exec(`INSERT INTO exercises 
			(topic_id, title, difficulty, description, initial_code, allowed_functions, test_cases, solution_hint) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			ex.topicID, ex.title, ex.difficulty, ex.desc, ex.initialCode, ex.allowedFunctions, ex.testCases, ex.hint)
		if err != nil {
			return err
		}
	}

	log.Println("Đã nạp thành công dữ liệu mẫu cho môn học.")
	return nil
}

// seedFunctions nạp hoặc cập nhật danh mục cú pháp cơ bản và hàm cho phép
func seedFunctions(db *sql.DB) error {
	funcs := []struct {
		name, category, syntax, desc, example string
	}{
		// Cú pháp cơ bản
		{"if / elif / else", "Cú pháp cơ bản", "if điều_kiện:\n    # khối lệnh\nelif điều_kiện_khác:\n    # khối lệnh\nelse:\n    # khối lệnh", "Cấu trúc rẽ nhánh điều kiện logic. Thực hiện khối lệnh khi điều kiện thỏa mãn True.", "score = 85\nif score >= 90:\n    print('A')\nelif score >= 80:\n    print('B')\nelse:\n    print('C')"},
		{"for", "Cú pháp cơ bản", "for biến in dãy_giá_trị:\n    # khối lệnh lặp", "Vòng lặp duyệt tuần tự từng phần tử trong một chuỗi, danh sách, hoặc kết quả hàm range().", "for x in [10, 20, 30]:\n    print(x)\nfor i in range(5):\n    print(i)"},
		{"while", "Cú pháp cơ bản", "while điều_kiện:\n    # khối lệnh lặp", "Vòng lặp tiếp tục thực thi chừng nào điều kiện logic vẫn còn thỏa mãn True.", "count = 3\nwhile count > 0:\n    print(count)\n    count -= 1"},
		{"def / return", "Cú pháp cơ bản", "def tên_hàm(tham_số):\n    # logic giải thuật\n    return giá_trị", "Định nghĩa một hàm xử lý và trả về kết quả bằng từ khóa return.", "def add(a, b):\n    return a + b\n\nres = add(3, 5) # 8"},
		{"break / continue", "Cú pháp cơ bản", "break # ngắt vòng lặp\ncontinue # sang vòng lặp kế tiếp", "break dùng để dừng và thoát hẳn khỏi vòng lặp. continue dùng để bỏ qua các lệnh còn lại của lượt lặp hiện tại.", "for i in range(10):\n    if i == 5:\n        break # dừng lặp\n    if i % 2 == 0:\n        continue # bỏ qua số chẵn\n    print(i)"},
		{"try / except", "Cú pháp cơ bản", "try:\n    # khối lệnh có thể lỗi\nexcept Lỗi:\n    # khối xử lý", "Bắt và xử lý lỗi/ngoại lệ trong thời gian thực thi (Runtime Error) mà không làm dừng chương trình.", "try:\n    x = 10 / 0\nexcept ZeroDivisionError:\n    print('Không thể chia cho 0')"},
		{"in / not in", "Cú pháp cơ bản", "giá_trị in tập_hợp\ngiá_trị not in tập_hợp", "Toán tử kiểm tra xem một phần tử có tồn tại trong danh sách (List), chuỗi (Str), hoặc khóa (Dict) hay không.", "nums = [1, 3, 5]\nif 3 in nums:\n    print('Tìm thấy 3')"},
		// Hàm có sẵn
		{"print", "Xuất dữ liệu", "print(*objects, sep=' ', end='\\n')", "In đối tượng hoặc giá trị ra màn hình console.", "print('Kết quả:', 42)"},
		{"len", "Cơ bản", "len(s)", "Trả về số lượng phần tử trong một chuỗi, danh sách hoặc từ điển.", "arr = [1, 2, 3]\nprint(len(arr)) # 3"},
		{"range", "Vòng lặp", "range(stop) hoặc range(start, stop, step)", "Tạo một dãy số nguyên tuần tự.", "for i in range(5):\n    print(i)"},
		{"append", "List", "list.append(x)", "Thêm một phần tử vào cuối danh sách.", "arr = [1]\narr.append(2)"},
		{"pop", "List", "list.pop([i])", "Xoá và trả về phần tử tại vị trí i (mặc định là phần tử cuối).", "arr = [10, 20]\nval = arr.pop() # 20"},
		{"insert", "List", "list.insert(i, x)", "Chèn một phần tử x vào vị trí chỉ số i.", "arr = [1, 3]\narr.insert(1, 2)"},
		{"int", "Ép kiểu", "int(x)", "Chuyển đổi một chuỗi hoặc số thực thành số nguyên.", "x = int('123')"},
		{"str", "Ép kiểu", "str(x)", "Chuyển đổi một đối tượng thành dạng chuỗi ký tự.", "s = str(123)"},
		{"min", "Toán học", "min(arg1, arg2, *args)", "Tìm giá trị nhỏ nhất trong các đối số hoặc trong danh sách.", "m = min(5, 2, 9)"},
		{"max", "Toán học", "max(arg1, arg2, *args)", "Tìm giá trị lớn nhất trong các đối số hoặc trong danh sách.", "m = max(5, 2, 9)"},
		{"sum", "Toán học", "sum(iterable)", "Tính tổng các phần tử trong danh sách.", "total = sum([1, 2, 3])"},
		{"dict.get", "Dict", "dict.get(key, default)", "Lấy giá trị theo khóa, trả về giá trị mặc định nếu không tồn tại.", "d = {'a': 1}\nprint(d.get('b', 0))"},
		{"dict.keys", "Dict", "dict.keys()", "Trả về danh sách các khóa trong bảng băm/từ điển.", "d = {'id': 1}\nkeys = d.keys()"},
		{"dict.values", "Dict", "dict.values()", "Trả về danh sách các giá trị trong từ điển.", "d = {'id': 1}\nvals = d.values()"},
	}

	for _, f := range funcs {
		_, err := db.Exec(`INSERT INTO functions (name, category, syntax, description, example) 
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET 
				category=excluded.category,
				syntax=excluded.syntax,
				description=excluded.description,
				example=excluded.example`,
			f.name, f.category, f.syntax, f.desc, f.example)
		if err != nil {
			return err
		}
	}
	return nil
}
