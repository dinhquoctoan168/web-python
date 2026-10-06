package judge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"web_python/internal/exercise"
)

var (
	ErrExerciseNotFound = errors.New("không tìm thấy bài tập cần chấm")
	ErrEmptySourceCode  = errors.New("mã nguồn nộp không được để trống")
)

type Service struct {
	exerciseService *exercise.Service
	pythonPath      string
	defaultTimeout  time.Duration
}

func NewService(es *exercise.Service) *Service {
	return &Service{
		exerciseService: es,
		pythonPath:      "python3",
		defaultTimeout:  3 * time.Second,
	}
}

func (s *Service) SetPythonPath(path string) {
	s.pythonPath = path
}

func (s *Service) SetDefaultTimeout(d time.Duration) {
	s.defaultTimeout = d
}

// Evaluate thực hiện chấm bài tập hoàn toàn tại server
func (s *Service) Evaluate(req JudgeRequest) (*JudgeResult, error) {
	if req.ExerciseID <= 0 {
		return nil, ErrExerciseNotFound
	}
	code := strings.TrimSpace(req.SourceCode)
	if code == "" {
		return nil, ErrEmptySourceCode
	}

	// 1. Tải đầy đủ thông tin bài tập bao gồm cả hidden test cases
	ex, err := s.exerciseService.GetExerciseForJudge(req.ExerciseID)
	if err != nil {
		return nil, fmt.Errorf("lỗi tải thông tin bài tập: %w", err)
	}
	if ex == nil {
		return nil, ErrExerciseNotFound
	}

	// 2. Kiểm tra an toàn mã nguồn
	if err := s.validateSecurity(code); err != nil {
		return &JudgeResult{
			Score:          0,
			PassedTests:    0,
			TotalTests:     len(ex.TestCases),
			Status:         "fail",
			ExecutionError: err.Error(),
		}, nil
	}

	// 3. Kiểm tra whitelist hàm (nếu bài tập có định nghĩa)
	if len(ex.AllowedFunctions) > 0 {
		if err := s.validateWhitelist(code, ex.AllowedFunctions); err != nil {
			return &JudgeResult{
				Score:          0,
				PassedTests:    0,
				TotalTests:     len(ex.TestCases),
				Status:         "fail",
				ExecutionError: err.Error(),
			}, nil
		}
	}

	// 4. Nếu bài tập không có test case định sẵn (ví dụ bài tập tự do)
	if len(ex.TestCases) == 0 {
		runRes := s.runSimpleScript(code)
		score := 0.0
		status := "fail"
		if runRes.Passed {
			score = 100.0
			status = "pass"
		}
		return &JudgeResult{
			Score:          score,
			PassedTests:    runRes.passInt(),
			TotalTests:     1,
			Status:         status,
			Tests:          []TestResult{runRes},
			ExecutionError: runRes.Error,
		}, nil
	}

	// 5. Chạy từng test case
	timeout := s.defaultTimeout
	if ex.TimeLimitMS > 0 {
		msTimeout := time.Duration(ex.TimeLimitMS) * time.Millisecond
		if msTimeout > 500*time.Millisecond && msTimeout <= 10*time.Second {
			timeout = msTimeout
		}
	}

	var rawResults []TestResult
	var totalWeight float64
	var passedWeight float64
	passCount := 0

	for _, tc := range ex.TestCases {
		w := tc.Weight
		if w <= 0 {
			w = 1.0
		}
		totalWeight += w

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		res := s.runTestCase(ctx, code, tc)
		cancel()

		if res.Passed {
			passedWeight += w
			passCount++
		}
		rawResults = append(rawResults, res)
	}

	// 6. Tính điểm số và trạng thái
	score := 0.0
	if totalWeight > 0 {
		score = (passedWeight / totalWeight) * 100.0
		// Làm tròn 2 chữ số thập phân
		score = float64(int(score*100+0.5)) / 100
	}

	status := "fail"
	if passCount == len(ex.TestCases) && len(ex.TestCases) > 0 {
		status = "pass"
	} else if passCount > 0 {
		status = "partial"
	}

	// 7. Bảo mật: Che giấu thông tin chi tiết của các test case ẩn
	var maskedResults []TestResult
	for _, r := range rawResults {
		if r.IsHidden {
			masked := r
			masked.Input = "[Test case ẩn]"
			masked.Call = "[Test case ẩn]"
			masked.Expected = "[Ẩn]"
			masked.Actual = "[Ẩn]"
			if masked.Error != "" {
				masked.Error = "Lỗi khi chạy test case ẩn"
			}
			maskedResults = append(maskedResults, masked)
		} else {
			maskedResults = append(maskedResults, r)
		}
	}

	return &JudgeResult{
		Score:       score,
		PassedTests: passCount,
		TotalTests:  len(ex.TestCases),
		Status:      status,
		Tests:       maskedResults,
	}, nil
}

func (s *Service) runTestCase(ctx context.Context, sourceCode string, tc exercise.TestCase) TestResult {
	start := time.Now()
	res := TestResult{
		TestCaseID: tc.ID,
		OrderNum:   tc.OrderNum,
		IsHidden:   tc.IsHidden,
		Weight:     tc.Weight,
		Input:      tc.InputData,
		Call:       tc.CallExpression,
		Expected:   tc.ExpectedOutput,
	}

	var script strings.Builder
	script.WriteString(sourceCode)
	script.WriteString("\n\n")

	var stdinInput string
	if tc.CallExpression != "" {
		script.WriteString("if __name__ == '__main__':\n")
		script.WriteString(fmt.Sprintf("    __res__ = %s\n", tc.CallExpression))
		script.WriteString("    print(__res__)\n")
	} else if tc.InputData != "" {
		stdinInput = tc.InputData
	}

	cmd := exec.CommandContext(ctx, s.pythonPath, "-c", script.String())
	if stdinInput != "" {
		cmd.Stdin = strings.NewReader(stdinInput)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()
	res.RuntimeMS = duration

	if ctx.Err() == context.DeadlineExceeded {
		res.Passed = false
		res.Error = "Quá thời gian thực thi (Time Limit Exceeded)"
		return res
	}

	if err != nil {
		res.Passed = false
		errMsg := strings.TrimSpace(stderrBuf.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		res.Error = errMsg
		res.Actual = strings.TrimSpace(stdoutBuf.String())
		return res
	}

	actual := strings.TrimSpace(stdoutBuf.String())
	res.Actual = actual

	if normalizeValue(tc.ExpectedOutput) == normalizeValue(actual) {
		res.Passed = true
	} else {
		res.Passed = false
	}

	return res
}

func (s *Service) runSimpleScript(sourceCode string) TestResult {
	start := time.Now()
	res := TestResult{
		OrderNum:  1,
		IsHidden:  false,
		Weight:    1.0,
		RuntimeMS: 0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.pythonPath, "-c", sourceCode)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	res.RuntimeMS = time.Since(start).Milliseconds()

	if ctx.Err() == context.DeadlineExceeded {
		res.Passed = false
		res.Error = "Quá thời gian thực thi"
		return res
	}

	if err != nil {
		res.Passed = false
		res.Error = strings.TrimSpace(stderrBuf.String())
		return res
	}

	res.Passed = true
	res.Actual = strings.TrimSpace(stdoutBuf.String())
	return res
}

func (s *Service) validateSecurity(code string) error {
	dangerous := []string{
		"import os", "import sys", "import subprocess", "import shutil",
		"import socket", "import pty", "import ctypes", "import signal",
		"from os", "from sys", "from subprocess", "from shutil",
		"from socket", "from pty", "from ctypes",
		"__import__", "eval(", "exec(", "open(",
	}

	for _, d := range dangerous {
		if strings.Contains(code, d) {
			return fmt.Errorf("phát hiện lệnh cấm '%s' vì lý do bảo mật máy chủ", d)
		}
	}
	return nil
}

func (s *Service) validateWhitelist(code string, allowed []string) error {
	allowedSet := make(map[string]bool)
	for _, fn := range allowed {
		allowedSet[strings.ToLower(strings.TrimSpace(fn))] = true
	}

	ignoredBuiltins := map[string]bool{
		"if": true, "while": true, "for": true, "elif": true, "return": true,
		"def": true, "class": true, "print": true, "len": true, "range": true,
		"str": true, "int": true, "float": true, "list": true, "dict": true,
		"set": true, "tuple": true, "bool": true, "abs": true, "min": true,
		"max": true, "sum": true, "sorted": true, "type": true, "isinstance": true,
	}

	re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	matches := re.FindAllStringSubmatch(code, -1)
	for _, m := range matches {
		if len(m) > 1 {
			fn := strings.ToLower(m[1])
			if ignoredBuiltins[fn] {
				continue
			}
			if !allowedSet[fn] {
				return fmt.Errorf("hàm '%s()' không nằm trong danh mục hàm được phép của bài tập", m[1])
			}
		}
	}
	return nil
}

func normalizeValue(val string) string {
	s := strings.TrimSpace(val)
	s = strings.ReplaceAll(s, "'", "\"")
	if strings.EqualFold(s, "true") {
		return "true"
	}
	if strings.EqualFold(s, "false") {
		return "false"
	}
	re := regexp.MustCompile(`\s*([,:[\](){}])\s*`)
	return re.ReplaceAllString(s, "$1")
}

func (t TestResult) passInt() int {
	if t.Passed {
		return 1
	}
	return 0
}
