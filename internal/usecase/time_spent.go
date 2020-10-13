package usecase

import (
	"context"
	"time"

	"github.com/pot-code/go-boilerplate/internal/domain"
	"go.elastic.co/apm"
)

// TimeSpentUseCase ...
type TimeSpentUseCase struct {
	TimeSpentRepository domain.TimeSpentRepository
}

// NewTimeSpentUseCase ...
func NewTimeSpentUseCase(
	TimeSpentRepository domain.TimeSpentRepository,
) *TimeSpentUseCase {
	return &TimeSpentUseCase{
		TimeSpentRepository: TimeSpentRepository,
	}
}

// GetUserTimeSpent get times spent on learning
//
// ts must be in RFC3339 layout
func (tsu *TimeSpentUseCase) GetUserTimeSpent(ctx context.Context, user *domain.UserModel, at *time.Time) ([]*domain.TimeSpentModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "TimeSpentUseCase.GetUserTimeSpent", "service")
	defer apmSpan.End()

	timeSpent, err := tsu.TimeSpentRepository.GetTimeSpentInWeekByUser(ctx, user, at)
	if err != nil {
		return nil, err
	}
	for _, e := range timeSpent {
		e.Timestamp = e.TS.Unix() * 1e3 // milliseconds
	}
	return timeSpent, nil
}
