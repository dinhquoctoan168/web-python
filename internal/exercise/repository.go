package exercise

import (
	"database/sql"
	"encoding/json"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindByID tìm bài tập theo id kèm thông tin chi tiết
func (r *Repository) FindByID(id int) (*Exercise, error) {
	query := `SELECT e.id, e.course_id, e.lesson_id, e.topic_id, COALESCE(t.name, ''), e.title, 
		e.exercise_type, e.difficulty, e.description, COALESCE(e.initial_code, ''), COALESCE(e.solution_code, ''), 
		COALESCE(e.solution_hint, ''), COALESCE(e.allowed_functions, '[]'), e.time_limit_ms, 
		COALESCE(e.visualization_type, ''), e.status, e.created_by, e.created_at
		FROM exercises e
		LEFT JOIN topics t ON e.topic_id = t.id
		WHERE e.id = ?`

	var ex Exercise
	var allowedFuncsRaw string
	var lessonID, topicID, createdBy sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&ex.ID, &ex.CourseID, &lessonID, &topicID, &ex.TopicName, &ex.Title,
		&ex.ExerciseType, &ex.Difficulty, &ex.Description, &ex.InitialCode, &ex.SolutionCode,
		&ex.SolutionHint, &allowedFuncsRaw, &ex.TimeLimitMS,
		&ex.VisualizationType, &ex.Status, &createdBy, &ex.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if lessonID.Valid {
		lid := int(lessonID.Int64)
		ex.LessonID = &lid
	}
	if topicID.Valid {
		tid := int(topicID.Int64)
		ex.TopicID = &tid
	}
	if createdBy.Valid {
		cid := int(createdBy.Int64)
		ex.CreatedBy = &cid
	}

	_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)
	return &ex, nil
}

// FindTestCasesByExerciseID lấy danh sách test cases, có cờ includeHidden để lấy hoặc lọc hidden test
func (r *Repository) FindTestCasesByExerciseID(exerciseID int, includeHidden bool) ([]TestCase, error) {
	var query string
	if includeHidden {
		query = `SELECT id, exercise_id, COALESCE(input_data, ''), COALESCE(call_expression, ''), 
			COALESCE(expected_output, ''), is_hidden, weight, order_num 
			FROM exercise_test_cases WHERE exercise_id = ? ORDER BY order_num ASC, id ASC`
	} else {
		query = `SELECT id, exercise_id, COALESCE(input_data, ''), COALESCE(call_expression, ''), 
			COALESCE(expected_output, ''), is_hidden, weight, order_num 
			FROM exercise_test_cases WHERE exercise_id = ? AND is_hidden = 0 ORDER BY order_num ASC, id ASC`
	}

	rows, err := r.db.Query(query, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []TestCase
	for rows.Next() {
		var tc TestCase
		var isHiddenInt int
		if err := rows.Scan(&tc.ID, &tc.ExerciseID, &tc.InputData, &tc.CallExpression, &tc.ExpectedOutput, &isHiddenInt, &tc.Weight, &tc.OrderNum); err != nil {
			return nil, err
		}
		tc.IsHidden = isHiddenInt == 1
		cases = append(cases, tc)
	}
	return cases, nil
}

// ListByCourseID lấy danh sách bài tập theo môn học
func (r *Repository) ListByCourseID(courseID int) ([]Exercise, error) {
	query := `SELECT id, course_id, lesson_id, topic_id, title, exercise_type, difficulty, 
		description, COALESCE(initial_code, ''), time_limit_ms, COALESCE(visualization_type, ''), status, created_at 
		FROM exercises WHERE course_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Exercise
	for rows.Next() {
		var ex Exercise
		var lessonID, topicID sql.NullInt64
		if err := rows.Scan(&ex.ID, &ex.CourseID, &lessonID, &topicID, &ex.Title, &ex.ExerciseType, &ex.Difficulty,
			&ex.Description, &ex.InitialCode, &ex.TimeLimitMS, &ex.VisualizationType, &ex.Status, &ex.CreatedAt); err != nil {
			return nil, err
		}
		if lessonID.Valid {
			lid := int(lessonID.Int64)
			ex.LessonID = &lid
		}
		if topicID.Valid {
			tid := int(topicID.Int64)
			ex.TopicID = &tid
		}
		list = append(list, ex)
	}
	return list, nil
}

// ListByLessonID lấy danh sách bài tập theo bài học
func (r *Repository) ListByLessonID(lessonID int) ([]Exercise, error) {
	query := `SELECT id, course_id, lesson_id, topic_id, title, exercise_type, difficulty, 
		description, COALESCE(initial_code, ''), time_limit_ms, COALESCE(visualization_type, ''), status, created_at 
		FROM exercises WHERE lesson_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(query, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Exercise
	for rows.Next() {
		var ex Exercise
		var lid, topicID sql.NullInt64
		if err := rows.Scan(&ex.ID, &ex.CourseID, &lid, &topicID, &ex.Title, &ex.ExerciseType, &ex.Difficulty,
			&ex.Description, &ex.InitialCode, &ex.TimeLimitMS, &ex.VisualizationType, &ex.Status, &ex.CreatedAt); err != nil {
			return nil, err
		}
		if lid.Valid {
			v := int(lid.Int64)
			ex.LessonID = &v
		}
		if topicID.Valid {
			tid := int(topicID.Int64)
			ex.TopicID = &tid
		}
		list = append(list, ex)
	}
	return list, nil
}

// ListAll lấy toàn bộ danh sách câu hỏi trong ngân hàng bài tập
func (r *Repository) ListAll() ([]Exercise, error) {
	query := `SELECT e.id, e.course_id, e.lesson_id, e.topic_id, COALESCE(t.name, ''), e.title, 
		e.exercise_type, e.difficulty, e.description, COALESCE(e.initial_code, ''), e.time_limit_ms, 
		COALESCE(e.visualization_type, ''), e.status, e.created_at 
		FROM exercises e 
		LEFT JOIN topics t ON e.topic_id = t.id 
		ORDER BY e.course_id ASC, e.id ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Exercise
	for rows.Next() {
		var ex Exercise
		var lid, tid sql.NullInt64
		if err := rows.Scan(&ex.ID, &ex.CourseID, &lid, &tid, &ex.TopicName, &ex.Title, &ex.ExerciseType, &ex.Difficulty,
			&ex.Description, &ex.InitialCode, &ex.TimeLimitMS, &ex.VisualizationType, &ex.Status, &ex.CreatedAt); err != nil {
			return nil, err
		}
		if lid.Valid {
			v := int(lid.Int64)
			ex.LessonID = &v
		}
		if tid.Valid {
			v := int(tid.Int64)
			ex.TopicID = &v
		}
		list = append(list, ex)
	}
	return list, nil
}

// GetCourseStructure lấy cấu trúc môn học gồm các chương và bài tập (an toàn cho client)
func (r *Repository) GetCourseStructure(courseID int) ([]ChapterWithExercises, error) {
	cRows, err := r.db.Query(`SELECT id, course_id, title, COALESCE(description, ''), order_num 
		FROM chapters WHERE course_id = ? ORDER BY order_num ASC, id ASC`, courseID)
	if err != nil {
		return nil, err
	}
	defer cRows.Close()

	var chapters []ChapterWithExercises
	for cRows.Next() {
		var c ChapterWithExercises
		if err := cRows.Scan(&c.ID, &c.CourseID, &c.Title, &c.Description, &c.OrderNum); err == nil {
			chapters = append(chapters, c)
		}
	}

	loadExercisesForQuery := func(query string, args ...any) ([]ClientExerciseDetail, error) {
		rows, err := r.db.Query(query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var exList []ClientExerciseDetail
		for rows.Next() {
			var ex ClientExerciseDetail
			var allowedFuncsRaw string
			var lid, tid sql.NullInt64
			if err := rows.Scan(
				&ex.ID, &ex.CourseID, &lid, &tid, &ex.TopicName, &ex.Title,
				&ex.ExerciseType, &ex.Difficulty, &ex.Description, &ex.InitialCode,
				&ex.SolutionHint, &allowedFuncsRaw, &ex.TimeLimitMS, &ex.VisualizationType,
			); err == nil {
				if lid.Valid {
					v := int(lid.Int64)
					ex.LessonID = &v
				}
				if tid.Valid {
					v := int(tid.Int64)
					ex.TopicID = &v
				}
				_ = json.Unmarshal([]byte(allowedFuncsRaw), &ex.AllowedFunctions)

				// Nạp public test cases
				tcs, tcErr := r.FindTestCasesByExerciseID(ex.ID, false)
				if tcErr == nil {
					var publicTCs []PublicTestCase
					for _, tc := range tcs {
						publicTCs = append(publicTCs, PublicTestCase{
							ID:             tc.ID,
							InputData:      tc.InputData,
							CallExpression: tc.CallExpression,
							ExpectedOutput: tc.ExpectedOutput,
							Weight:         tc.Weight,
							OrderNum:       tc.OrderNum,
						})
					}
					ex.TestCases = publicTCs
					if tcsJSON, jErr := json.Marshal(publicTCs); jErr == nil {
						ex.TestCasesJSON = string(tcsJSON)
					} else {
						ex.TestCasesJSON = "[]"
					}
				} else {
					ex.TestCasesJSON = "[]"
				}

				exList = append(exList, ex)
			}
		}
		return exList, nil
	}

	exSelectSQL := `SELECT e.id, e.course_id, e.lesson_id, e.topic_id, COALESCE(t.name, ''), e.title, 
		e.exercise_type, e.difficulty, e.description, COALESCE(e.initial_code, ''), 
		COALESCE(e.solution_hint, ''), COALESCE(e.allowed_functions, '[]'), e.time_limit_ms, 
		COALESCE(e.visualization_type, '')
		FROM exercises e
		LEFT JOIN lessons l ON e.lesson_id = l.id
		LEFT JOIN topics t ON e.topic_id = t.id`

	if len(chapters) == 0 {
		// Trường hợp môn học chưa chia chương, gom toàn bộ bài tập vào 1 nhóm mặc định
		exercises, err := loadExercisesForQuery(exSelectSQL+` WHERE e.course_id = ? ORDER BY e.id ASC`, courseID)
		if err != nil {
			return nil, err
		}
		if len(exercises) > 0 {
			chapters = append(chapters, ChapterWithExercises{
				ID:          0,
				CourseID:    courseID,
				Title:       "Danh sách bài tập",
				Description: "",
				Exercises:   exercises,
			})
		}
		return chapters, nil
	}

	for i := range chapters {
		chapID := chapters[i].ID
		var exercises []ClientExerciseDetail
		if i == 0 {
			// Nhóm đầu tiên gom thêm các bài tập thuộc course mà lesson_id is null
			exercises, _ = loadExercisesForQuery(exSelectSQL+` WHERE l.chapter_id = ? OR (e.course_id = ? AND e.lesson_id IS NULL) ORDER BY e.id ASC`, chapID, courseID)
		} else {
			exercises, _ = loadExercisesForQuery(exSelectSQL+` WHERE l.chapter_id = ? ORDER BY e.id ASC`, chapID)
		}
		chapters[i].Exercises = exercises
	}

	return chapters, nil
}
