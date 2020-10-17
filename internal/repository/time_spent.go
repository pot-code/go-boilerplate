package repository

import (
	"context"
	"time"

	"github.com/pot-code/go-boilerplate/internal/domain"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
)

type TimeSpentRepository struct {
	Conn driver.ITransactionalDB `dep:""`
}

func NewTimeSpentRepository(Conn driver.ITransactionalDB) *TimeSpentRepository {
	return &TimeSpentRepository{
		Conn: Conn,
	}
}

func (repo *TimeSpentRepository) GetTimeSpentInWeekByUser(ctx context.Context, user *domain.UserModel, at *time.Time) ([]*domain.TimeSpentModel, error) {
	conn := repo.Conn
	rows, err := conn.QueryContext(ctx, `
SELECT 
    WEEKDAY(ts),
    SUM(vocabulary) vocabulary,
    SUM(grammar) grammar,
    SUM(listening) listening,
    SUM(writing) writing,
    ts
FROM
    lesson_time_spent
WHERE
    YEARWEEK(ts, 1) = YEARWEEK($1, 1)
        AND user_id = $2
GROUP BY ts
ORDER BY ts ASC;
	`, at, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.TimeSpentModel
	for rows.Next() {
		item := new(domain.TimeSpentModel)
		err := rows.Scan(&item.Weekday, &item.Vocabulary, &item.Grammar, &item.Listening, &item.Writing, &item.TS)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
