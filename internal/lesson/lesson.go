package lesson

import (
	"context"
	"time"

	"github.com/pot-code/go-boilerplate/internal/user"
)

type LessonProgressModel struct {
	ID        int        `json:"id"`
	Index     int        `json:"index"`
	Title     string     `json:"title"`
	Progress  float32    `json:"progress"`
	CreatedAt *time.Time `json:"-"`
	Timestamp int64      `json:"timestamp"`
}

type LessonRepository interface {
	GetLessonProgressByUser(ctx context.Context, user *user.UserModel) ([]*LessonProgressModel, error)
}

type LessonUseCase interface {
	GetUserLessonProgress(ctx context.Context, user *user.UserModel) ([]*LessonProgressModel, error)
}
