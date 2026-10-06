/**
 * AlgoDB Python Execution Engine
 * Thực thi mã nguồn Python và kiểm duyệt giới hạn hàm (Whitelist Validator).
 */

const PythonEngine = {
    // Các từ khóa và built-ins luôn an toàn cho giải thuật
    BASE_SAFE_BUILTINS: [
        'print', 'len', 'range', 'int', 'float', 'str', 'bool', 'list', 'dict', 'set', 'tuple'
    ],

    // Các hàm/từ khóa cấm hoàn toàn vì lý do bảo mật hoặc gian lận
    STRICT_FORBIDDEN: [
        'eval', 'exec', 'open', '__import__', 'compile', 'globals', 'locals', 'vars', 'input'
    ],

    /**
     * Kiểm tra mã nguồn có vi phạm giới hạn hàm (whitelist) của bài tập hay không.
     * @param {string} code - Mã nguồn Python của sinh viên
     * @param {string[]} allowedFunctions - Danh sách hàm cho phép theo bài tập
     * @returns {{isValid: boolean, errorMsg: string|null}}
     */
    validateCode: function(code, allowedFunctions) {
        if (!code || code.trim() === '') {
            return { isValid: false, errorMsg: 'Mã nguồn rỗng. Vui lòng nhập mã trước khi chạy.' };
        }

        // 1. Kiểm tra các hàm tuyệt đối cấm
        for (const forbidden of this.STRICT_FORBIDDEN) {
            const regex = new RegExp(`\\b${forbidden}\\s*\\(`, 'g');
            if (regex.test(code)) {
                return {
                    isValid: false,
                    errorMsg: `CẢNH BÁO BẢO MẬT: Hàm '${forbidden}()' bị cấm tuyệt đối trong môi trường học tập.`
                };
            }
        }

        // 2. Kiểm tra nếu có giới hạn hàm cụ thể cho bài tập
        if (Array.isArray(allowedFunctions) && allowedFunctions.length > 0) {
            // Chuẩn hóa danh sách cho phép
            const allowedSet = new Set(allowedFunctions.map(f => f.trim().toLowerCase()));

            // Tìm các lệnh gọi hàm dạng `func_name(`
            const funcCallRegex = /([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)\s*\(/g;
            let match;
            const calledFuncs = [];

            // Loại bỏ chuỗi và comment để tránh false positive
            const sanitizedCode = code
                .replace(/#.*$/gm, '')              // bỏ comment 1 dòng
                .replace(/'''[\s\S]*?'''/g, '')     // bỏ docstring nháy đơn
                .replace(/"""[\s\S]*?"""/g, '');    // bỏ docstring nháy kép

            // Lấy danh sách hàm do sinh viên tự định nghĩa (def my_func)
            const userDefinedFuncs = new Set();
            const defRegex = /def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(/g;
            let defMatch;
            while ((defMatch = defRegex.exec(sanitizedCode)) !== null) {
                userDefinedFuncs.add(defMatch[1].toLowerCase());
            }

            while ((match = funcCallRegex.exec(sanitizedCode)) !== null) {
                const rawName = match[1];
                const funcName = rawName.toLowerCase();

                // Bỏ qua hàm do chính người dùng định nghĩa
                if (userDefinedFuncs.has(funcName)) {
                    continue;
                }

                // Bỏ qua các cấu trúc rẽ nhánh / vòng lặp vô tình khớp
                if (['if', 'while', 'for', 'elif', 'return'].includes(funcName)) {
                    continue;
                }

                calledFuncs.push(rawName);
            }

            // Đối chiếu hàm đã gọi với Whitelist
            for (const called of calledFuncs) {
                const calledLower = called.toLowerCase();
                const shortName = calledLower.includes('.') ? calledLower.split('.').pop() : calledLower;

                const isAllowed = allowedSet.has(calledLower) || 
                                  allowedSet.has(shortName) || 
                                  allowedSet.has(called);

                if (!isAllowed) {
                    return {
                        isValid: false,
                        errorMsg: `GIỚI HẠN BÀI HỌC: Bạn đang gọi hàm '${called}()'.\n` +
                                  `Hàm này không nằm trong danh mục hàm được phép của bài tập!\n` +
                                  `Các hàm được phép sử dụng: [ ${Array.from(allowedSet).join(', ')} ]`
                    };
                }
            }
        }

        return { isValid: true, errorMsg: null };
    },

    /**
     * Khởi tạo môi trường và thực thi mã nguồn Python.
     * @param {string} code - Mã Python
     * @param {function} onOutput - Callback nhận từng dòng stdout
     * @param {function} onError - Callback khi có lỗi biên dịch hoặc runtime
     * @param {function} onComplete - Callback khi hoàn thành
     */
    runPython: function(code, onOutput, onError, onComplete) {
        if (typeof Sk === 'undefined') {
            onError('Lỗi hệ thống: Môi trường thực thi chưa sẵn sàng. Vui lòng làm mới trang.');
            if (onComplete) onComplete(false, 0);
            return;
        }

        const startTime = performance.now();

        // Cấu hình môi trường Skulpt
        Sk.configure({
            output: function(text) {
                if (onOutput) onOutput(text);
            },
            read: function(x) {
                if (Sk.builtinFiles === undefined || Sk.builtinFiles["files"][x] === undefined) {
                    throw "Không tìm thấy thư viện: '" + x + "'";
                }
                return Sk.builtinFiles["files"][x];
            },
            __future__: Sk.python3
        });

        // Thiết lập timeout ngắt lặp vô tận (Execution limit)
        Sk.execLimit = 5000; // 5 giây tối đa

        const promise = Sk.misceval.asyncToPromise(function() {
            return Sk.importMainWithBody("<stdin>", false, code, true);
        });

        promise.then(
            function(mod) {
                const duration = Math.round(performance.now() - startTime);
                if (onComplete) onComplete(true, duration);
            },
            function(err) {
                const duration = Math.round(performance.now() - startTime);
                let errString = err.toString();
                if (err.traceback && err.traceback.length > 0) {
                    const tb = err.traceback[0];
                    errString += `\n  tại dòng ${tb.lineno}`;
                }
                if (onError) onError(errString);
                if (onComplete) onComplete(false, duration);
            }
        );
    },

    /**
     * Chuẩn hóa giá trị để so khớp kết quả giữa Actual và Expected
     */
    normalizeValue: function(val) {
        if (val === null || val === undefined) return '';
        let s = String(val).trim();
        // Thay thế nháy đơn thành nháy kép
        s = s.replace(/'/g, '"');
        // Chuẩn hóa khoảng trắng quanh các ký tự cấu trúc
        s = s.replace(/\s*([,:[\](){}])\s*/g, '$1');
        // Chuẩn hóa boolean
        if (s.toLowerCase() === 'true') return 'true';
        if (s.toLowerCase() === 'false') return 'false';
        return s;
    },

    /**
     * Chạy toàn bộ bộ test cases kiểm thử bài tập
     * Thực thi hàm của sinh viên độc lập cho từng bộ dữ liệu đầu vào.
     * @param {string} code - Mã nguồn sinh viên
     * @param {Array} testCases - Danh sách test cases
     * @param {function} onCaseFinished - Callback cập nhật từng case
     * @param {function} onAllFinished - Callback hoàn tất toàn bộ
     */
    runTests: function(code, testCases, onCaseFinished, onAllFinished) {
        if (!testCases || testCases.length === 0) {
            if (onAllFinished) onAllFinished([], 0, 0);
            return;
        }

        // Tìm tên hàm chính được định nghĩa trong code sinh viên
        const funcMatch = code.match(/def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(/);
        const mainFunc = funcMatch ? funcMatch[1] : null;

        // Xây dựng đoạn mã test harness độc lập
        let harness = "\n\n# --- AUTOMATED TEST RUNNER ---\n";
        harness += 'print("__BEGIN_TEST_HARNESS__")\n';

        testCases.forEach((tc, idx) => {
            let callExpr = tc.call || (mainFunc && tc.input ? `${mainFunc}(${tc.input})` : null);
            if (callExpr) {
                harness += `try:\n`;
                harness += `    __res_${idx}__ = ${callExpr}\n`;
                harness += `    print("__TC_VAL__${idx}:" + str(__res_${idx}__))\n`;
                harness += `except Exception as __e_${idx}__:\n`;
                harness += `    print("__TC_ERR__${idx}:" + str(__e_${idx}__))\n`;
            } else {
                harness += `print("__TC_ERR__${idx}:Không xác định được hàm kiểm thử")\n`;
            }
        });

        harness += 'print("__END_TEST_HARNESS__")\n';

        const testExecutionCode = code + harness;

        let capturedOutput = "";
        this.runPython(
            testExecutionCode,
            function(out) {
                capturedOutput += out;
            },
            function(err) {
                capturedOutput += "\n[Lỗi Runtime]: " + err;
            },
            function(success, duration) {
                const results = [];
                let passedCount = 0;

                // Phân tích output giữa __BEGIN_TEST_HARNESS__ và __END_TEST_HARNESS__
                const beginMarker = "__BEGIN_TEST_HARNESS__";
                const endMarker = "__END_TEST_HARNESS__";
                const beginIdx = capturedOutput.indexOf(beginMarker);
                const endIdx = capturedOutput.indexOf(endMarker);

                let harnessOutput = "";
                if (beginIdx !== -1 && endIdx !== -1) {
                    harnessOutput = capturedOutput.substring(beginIdx + beginMarker.length, endIdx);
                } else {
                    harnessOutput = capturedOutput;
                }

                testCases.forEach((tc, idx) => {
                    const valRegex = new RegExp(`__TC_VAL__${idx}:([\\s\\S]*?)(?=\\n__TC_|$|\\n)`);
                    const errRegex = new RegExp(`__TC_ERR__${idx}:([\\s\\S]*?)(?=\\n__TC_|$|\\n)`);

                    const valMatch = harnessOutput.match(valRegex);
                    const errMatch = harnessOutput.match(errRegex);

                    let actual = "";
                    let isPass = false;

                    if (errMatch) {
                        actual = `Lỗi thực thi: ${errMatch[1].trim()}`;
                        isPass = false;
                    } else if (valMatch) {
                        actual = valMatch[1].trim();
                        const normActual = PythonEngine.normalizeValue(actual);
                        const normExpected = PythonEngine.normalizeValue(tc.expected);
                        isPass = (normActual === normExpected);
                    } else {
                        // Fallback nếu code sinh viên bị lỗi chặn trước khi chạy tới harness
                        actual = "Không nhận được giá trị trả về (Lỗi biên dịch hoặc lỗi cú pháp)";
                        isPass = false;
                    }

                    if (isPass) passedCount++;

                    const resultItem = {
                        index: idx + 1,
                        input: tc.input || tc.call || "",
                        expected: tc.expected,
                        actual: actual,
                        passed: isPass
                    };

                    results.push(resultItem);
                    if (onCaseFinished) onCaseFinished(resultItem);
                });

                if (onAllFinished) onAllFinished(results, passedCount, testCases.length);
            }
        );
    }
};

window.PythonEngine = PythonEngine;
