package lesson

import (
	"context"

	"github.com/pot-code/go-boilerplate/internal/domain"
	"go.elastic.co/apm"
)

// LessonUseCase ...
type LessonUseCase struct {
	LessonRepository domain.LessonRepository
}

// NewLessonUseCase ...
func NewLessonUseCase(
	LessonRepository domain.LessonRepository,
) *LessonUseCase {
	return &LessonUseCase{LessonRepository}
}

// GetUserLessonProgress get learning progress for each lesson
func (lu *LessonUseCase) GetUserLessonProgress(ctx context.Context, user *domain.UserModel) ([]*domain.LessonProgressModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "LessonUseCase.GetUserLessonProgress", "service")
	defer apmSpan.End()

	progress, err := lu.LessonRepository.GetLessonProgressByUser(ctx, user)
	if err != nil {
		return nil, err
	}
	for _, e := range progress {
		e.Timestamp = e.CreatedAt.Unix() * 1e3 // milliseconds
	}
	return progress, nil
}
