package lesson

import (
	"context"

	"github.com/pot-code/go-boilerplate/internal/user"
	"go.elastic.co/apm"
)

// LessonUseCaseImpl ...
type LessonUseCaseImpl struct {
	LessonRepository LessonRepository
}

var _ LessonUseCase = &LessonUseCaseImpl{}

// NewLessonUseCase ...
func NewLessonUseCase(
	LessonRepository LessonRepository,
) *LessonUseCaseImpl {
	return &LessonUseCaseImpl{LessonRepository}
}

// GetUserLessonProgress get learning progress for each lesson
func (lu *LessonUseCaseImpl) GetUserLessonProgress(ctx context.Context, user *user.UserModel) ([]*LessonProgressModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "LessonUseCaseImpl.GetUserLessonProgress", "service")
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
