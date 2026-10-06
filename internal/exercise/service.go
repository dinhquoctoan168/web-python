package exercise

import (
	"encoding/json"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetExerciseForClient lấy thông tin bài tập trả về client/browser
// TUYỆT ĐỐI không bao gồm test cases ẩn (is_hidden = 1) và không bao gồm solution_code
func (s *Service) GetExerciseForClient(id int) (*ClientExerciseDetail, error) {
	ex, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("không thể tìm bài tập: %w", err)
	}
	if ex == nil {
		return nil, nil
	}

	// Chỉ lấy public test cases (includeHidden = false)
	testCases, err := s.repo.FindTestCasesByExerciseID(id, false)
	if err != nil {
		return nil, fmt.Errorf("không thể tải test cases: %w", err)
	}

	var publicTCs []PublicTestCase
	for _, tc := range testCases {
		if !tc.IsHidden {
			publicTCs = append(publicTCs, PublicTestCase{
				ID:             tc.ID,
				InputData:      tc.InputData,
				CallExpression: tc.CallExpression,
				ExpectedOutput: tc.ExpectedOutput,
				Weight:         tc.Weight,
				OrderNum:       tc.OrderNum,
			})
		}
	}

	tcsJSON, err := json.Marshal(publicTCs)
	if err != nil {
		return nil, fmt.Errorf("lỗi đóng gói JSON test cases: %w", err)
	}

	return &ClientExerciseDetail{
		ID:               ex.ID,
		CourseID:         ex.CourseID,
		LessonID:         ex.LessonID,
		TopicID:          ex.TopicID,
		TopicName:        ex.TopicName,
		Title:            ex.Title,
		ExerciseType:     ex.ExerciseType,
		Difficulty:       ex.Difficulty,
		Description:      ex.Description,
		InitialCode:      ex.InitialCode,
		SolutionHint:     ex.SolutionHint,
		AllowedFunctions:  ex.AllowedFunctions,
		TimeLimitMS:       ex.TimeLimitMS,
		VisualizationType: ex.VisualizationType,
		TestCases:         publicTCs,
		TestCasesJSON:     string(tcsJSON),
	}, nil
}

// GetExerciseForJudge lấy toàn bộ thông tin bài tập gồm cả hidden test cases và solution_code cho server chấm bài
func (s *Service) GetExerciseForJudge(id int) (*Exercise, error) {
	ex, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("không thể tìm bài tập: %w", err)
	}
	if ex == nil {
		return nil, nil
	}

	// Lấy tất cả test cases kể cả hidden
	testCases, err := s.repo.FindTestCasesByExerciseID(id, true)
	if err != nil {
		return nil, fmt.Errorf("không thể tải test cases: %w", err)
	}
	ex.TestCases = testCases

	return ex, nil
}

// ListAll lấy toàn bộ danh sách câu hỏi trong ngân hàng bài tập
func (s *Service) ListAll() ([]Exercise, error) {
	return s.repo.ListAll()
}
