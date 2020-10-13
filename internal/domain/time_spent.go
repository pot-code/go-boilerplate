package domain

import (
	"context"
	"time"
)

type TimeSpentModel struct {
	ID         int        `json:"-"`
	UserID     string     `json:"-"`
	Vocabulary int        `json:"vocabulary"`
	Grammar    int        `json:"grammar"`
	Listening  int        `json:"listening"`
	Writing    int        `json:"writing"`
	TS         *time.Time `json:"-"`
	Weekday    int        `json:"weekday"`
	Timestamp  int64      `json:"timestamp"`
}

type TimeSpentRepository interface {
	GetTimeSpentInWeekByUser(ctx context.Context, user *UserModel, at *time.Time) ([]*TimeSpentModel, error)
}

type TimeSpentUseCase interface {
	GetUserTimeSpent(ctx context.Context, user *UserModel, until *time.Time) ([]*TimeSpentModel, error)
}
