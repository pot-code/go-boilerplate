package timespent

import (
	"context"
	"time"

	"github.com/pot-code/go-boilerplate/internal/user"
	"go.elastic.co/apm"
)

// TimeSpentUseCaseImpl ...
type TimeSpentUseCaseImpl struct {
	TimeSpentRepository TimeSpentRepository
}

var _ TimeSpentUseCase = &TimeSpentUseCaseImpl{}

// NewTimeSpentUseCase ...
func NewTimeSpentUseCase(
	TimeSpentRepository TimeSpentRepository,
) *TimeSpentUseCaseImpl {
	return &TimeSpentUseCaseImpl{TimeSpentRepository}
}

// GetUserTimeSpent get times spent on learning
//
// ts must be in RFC3339 layout
func (tsu *TimeSpentUseCaseImpl) GetUserTimeSpent(ctx context.Context, user *user.UserModel, at *time.Time) ([]*TimeSpentModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "TimeSpentUseCaseImpl.GetUserTimeSpent", "service")
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
