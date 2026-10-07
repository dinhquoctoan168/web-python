package judge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

var (
	skulptInitOnce    sync.Once
	skulptInitErr     error
	compiledMock      *goja.Program
	compiledSkulpt    *goja.Program
	compiledStdlib    *goja.Program
)

func resolveAssetPath(relPath string) string {
	if _, err := os.Stat(relPath); err == nil {
		return relPath
	}
	alt := filepath.Join("..", "..", relPath)
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return relPath
}

// initSkulptPrograms biên dịch sẵn các file JavaScript của Skulpt một lần duy nhất
func initSkulptPrograms() error {
	skulptInitOnce.Do(func() {
		mockCode := `
			var window = this;
			var global = this;
			var console = { log: function() {}, error: function() {}, warn: function() {} };
		`
		var err error
		compiledMock, err = goja.Compile("mock.js", mockCode, true)
		if err != nil {
			skulptInitErr = fmt.Errorf("biên dịch mock thất bại: %w", err)
			return
		}

		skulptPath := resolveAssetPath("web/static/js/skulpt.min.js")
		skulptBytes, err := os.ReadFile(skulptPath)
		if err != nil {
			skulptInitErr = fmt.Errorf("không thể đọc %s: %w", skulptPath, err)
			return
		}

		compiledSkulpt, err = goja.Compile("skulpt.min.js", string(skulptBytes), true)
		if err != nil {
			skulptInitErr = fmt.Errorf("biên dịch skulpt.min.js thất bại: %w", err)
			return
		}

		stdlibPath := resolveAssetPath("web/static/js/skulpt-stdlib.js")
		stdlibBytes, err := os.ReadFile(stdlibPath)
		if err != nil {
			skulptInitErr = fmt.Errorf("không thể đọc %s: %w", stdlibPath, err)
			return
		}

		compiledStdlib, err = goja.Compile("skulpt-stdlib.js", string(stdlibBytes), true)
		if err != nil {
			skulptInitErr = fmt.Errorf("biên dịch skulpt-stdlib.js thất bại: %w", err)
			return
		}
	})

	return skulptInitErr
}

// SkulptExecutionResult lưu kết quả chạy code Python bằng Skulpt
type SkulptExecutionResult struct {
	Output    string
	Error     string
	RuntimeMS int64
	TimedOut  bool
}

// ExecutePythonWithSkulpt thực thi mã Python hoàn toàn trong bộ nhớ bằng engine Skulpt
func ExecutePythonWithSkulpt(code string, inputData string, timeout time.Duration) SkulptExecutionResult {
	start := time.Now()
	res := SkulptExecutionResult{}

	if err := initSkulptPrograms(); err != nil {
		res.Error = "Lỗi khởi tạo môi trường Skulpt: " + err.Error()
		return res
	}

	vm := goja.New()

	// Thiết lập giới hạn thời gian thực thi chống vòng lặp vô hạn
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	timer := time.AfterFunc(timeout, func() {
		vm.Interrupt("Quá thời gian thực thi (Time Limit Exceeded)")
	})
	defer timer.Stop()

	if _, err := vm.RunProgram(compiledMock); err != nil {
		res.Error = "Lỗi nạp mock: " + err.Error()
		return res
	}
	if _, err := vm.RunProgram(compiledSkulpt); err != nil {
		res.Error = "Lỗi nạp Skulpt: " + err.Error()
		return res
	}
	if _, err := vm.RunProgram(compiledStdlib); err != nil {
		res.Error = "Lỗi nạp Stdlib: " + err.Error()
		return res
	}

	var output strings.Builder
	vm.Set("__onOutput", func(call goja.FunctionCall) goja.Value {
		output.WriteString(call.Argument(0).String())
		return goja.Undefined()
	})

	// Chuẩn bị danh sách input theo dòng nếu có
	inputLines := strings.Split(inputData, "\n")
	vm.Set("__inputLines", inputLines)
	vm.Set("__inputIndex", 0)

	jsRunner := `
		var __execError = null;
		Sk.configure({
			output: __onOutput,
			read: function(x) {
				if (Sk.builtinFiles === undefined || Sk.builtinFiles["files"][x] === undefined) {
					throw "Không tìm thấy thư viện: " + x;
				}
				return Sk.builtinFiles["files"][x];
			},
			inputfun: function() {
				if (__inputIndex < __inputLines.length) {
					var line = __inputLines[__inputIndex++];
					return line;
				}
				return "";
			},
			inputfunTakesPrompt: true,
			__future__: Sk.python3
		});

		try {
			Sk.importMainWithBody("<stdin>", false, __sourceCode, false);
		} catch (e) {
			__execError = e.toString();
		}
	`

	vm.Set("__sourceCode", code)

	_, err := vm.RunString(jsRunner)
	res.RuntimeMS = time.Since(start).Milliseconds()

	if err != nil {
		var interruptedErr *goja.InterruptedError
		if errors.As(err, &interruptedErr) || strings.Contains(err.Error(), "Time Limit Exceeded") {
			res.TimedOut = true
			res.Error = "Quá thời gian thực thi (Time Limit Exceeded)"
		} else {
			res.Error = err.Error()
		}
		res.Output = strings.TrimSpace(output.String())
		return res
	}

	if execErrVal := vm.Get("__execError"); execErrVal != nil && !goja.IsNull(execErrVal) && !goja.IsUndefined(execErrVal) {
		res.Error = execErrVal.String()
	}

	res.Output = strings.TrimSpace(output.String())
	return res
}
