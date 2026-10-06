package progress

// ExerciseProgressItem biểu diễn trạng thái làm bài tập cụ thể của học viên
type ExerciseProgressItem struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"exercise_type"`
	Status    string `json:"status"` // "not_started", "in_progress", "completed"
	BestScore int    `json:"best_score"`
}

// LessonProgress biểu diễn tiến độ của một bài học
type LessonProgress struct {
	LessonID           int                    `json:"lesson_id"`
	ChapterID          int                    `json:"chapter_id"`
	Title              string                 `json:"title"`
	TotalExercises     int                    `json:"total_exercises"`
	CompletedExercises int                    `json:"completed_exercises"`
	ProgressPercent    int                    `json:"progress_percent"`
	IsCompleted        bool                   `json:"is_completed"`
	Exercises          []ExerciseProgressItem `json:"exercises,omitempty"`
}

// ChapterProgress biểu diễn tiến độ của một chương
type ChapterProgress struct {
	ChapterID          int              `json:"chapter_id"`
	CourseID           int              `json:"course_id"`
	Title              string           `json:"title"`
	TotalLessons       int              `json:"total_lessons"`
	CompletedLessons   int              `json:"completed_lessons"`
	ProgressPercent    int              `json:"progress_percent"`
	IsCompleted        bool             `json:"is_completed"`
	TotalExercises     int              `json:"total_exercises"`
	CompletedExercises int              `json:"completed_exercises"`
	Lessons            []LessonProgress `json:"lessons,omitempty"`
}

// CourseProgress biểu diễn tiến độ tổng thể của một môn học
type CourseProgress struct {
	CourseID            int               `json:"course_id"`
	CourseName          string            `json:"course_name"`
	TotalActivities     int               `json:"total_activities"` // Tổng số bài tập/hoạt động yêu cầu
	CompletedActivities int               `json:"completed_activities"`
	ProgressPercent     int               `json:"progress_percent"`
	TotalLessons        int               `json:"total_lessons"`
	CompletedLessons    int               `json:"completed_lessons"`
	IsCompleted         bool              `json:"is_completed"`
	Chapters            []ChapterProgress `json:"chapters,omitempty"`
}
