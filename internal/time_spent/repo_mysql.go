package timespent

import (
	"context"
	"time"

	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/user"
)

type TimeSpentMySQL struct {
	Conn driver.ITransactionalDB `dep:""`
}

var _ TimeSpentRepository = &TimeSpentMySQL{}

func NewTimeSpentRepository(Conn driver.ITransactionalDB) *TimeSpentMySQL {
	return &TimeSpentMySQL{
		Conn: Conn,
	}
}

func (repo *TimeSpentMySQL) GetTimeSpentInWeekByUser(ctx context.Context, user *user.UserModel, at *time.Time) ([]*TimeSpentModel, error) {
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

	var result []*TimeSpentModel
	for rows.Next() {
		item := new(TimeSpentModel)
		err := rows.Scan(&item.Weekday, &item.Vocabulary, &item.Grammar, &item.Listening, &item.Writing, &item.TS)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
