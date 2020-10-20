package lesson

import (
	"context"

	"github.com/pot-code/go-boilerplate/internal/domain"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
)

type LessonRepository struct {
	Conn driver.ITransactionalDB `dep:""`
}

var _ domain.LessonRepository = &LessonRepository{}

func NewLessonRepository(Conn driver.ITransactionalDB) *LessonRepository {
	return &LessonRepository{
		Conn: Conn,
	}
}

func (repo *LessonRepository) GetLessonProgressByUser(ctx context.Context, user *domain.UserModel) ([]*domain.LessonProgressModel, error) {
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

	var result []*domain.LessonProgressModel
	for rows.Next() {
		item := new(domain.LessonProgressModel)
		err := rows.Scan(&item.ID, &item.Index, &item.Title, &item.Progress, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
