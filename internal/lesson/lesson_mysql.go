package lesson

import (
	"context"

	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/user"
)

type LessonMySQL struct {
	Conn driver.ITransactionalDB `dep:""`
}

var _ LessonRepository = &LessonMySQL{}

func NewLessonRepository(Conn driver.ITransactionalDB) *LessonMySQL {
	return &LessonMySQL{
		Conn: Conn,
	}
}

func (repo *LessonMySQL) GetLessonProgressByUser(ctx context.Context, user *user.UserModel) ([]*LessonProgressModel, error) {
	conn := repo.Conn
	rows, err := conn.QueryContext(ctx, `
SELECT 
    lp.id, l."index", l."name" title, lp.progress, lp.created_at
FROM
    lesson_progress lp
        LEFT JOIN
    lesson l ON (l.id = lp.lesson_id)
WHERE
    lp.user_id = $1
	`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*LessonProgressModel
	for rows.Next() {
		item := new(LessonProgressModel)
		err := rows.Scan(&item.ID, &item.Index, &item.Title, &item.Progress, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
