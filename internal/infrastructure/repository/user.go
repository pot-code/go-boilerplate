package repo

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/pot-code/go-boilerplate/internal/domain"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
)

type UserRepository struct {
	Conn driver.ITransactionalDB
	Tx   driver.ITransactionalDB
}

func NewUserRepository(Conn driver.ITransactionalDB) *UserRepository {
	return &UserRepository{
		Conn: Conn,
	}
}

// FindByCredential query user with provided credential
func (repo *UserRepository) FindByCredential(ctx context.Context, post *domain.UserModel) (*domain.UserModel, error) {
	conn := repo.Conn
	username := post.Username
	email := post.Email
	row, err := conn.QueryContext(ctx, `SELECT id, username, password, email
	FROM user WHERE username=? OR email=?`, username, email)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		user := new(domain.UserModel)
		if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email); err != nil {
			return nil, err
		}
		return user, nil
	}
	return nil, nil
}

func (repo *UserRepository) SaveUser(ctx context.Context, post *domain.UserModel) error {
	conn := repo.Conn
	_, err := conn.ExecContext(ctx, `INSERT INTO user(id, username, password, email)
	VALUES(?,?,?,?)`, post.ID, post.Username, post.Password, post.Email)

	if err, ok := err.(*mysql.MySQLError); ok && err.Number == 1062 {
		return domain.ErrDuplicatedUser
	}
	return err
}

func (repo *UserRepository) BeginTx(ctx context.Context) error {
	if repo.Tx != nil {
		return nil
	}
	if tx, err := repo.Conn.BeginTx(ctx, &driver.TxOptions{
		Isolation: sql.LevelReadCommitted,
	}); err == nil {
		repo.Tx = tx
	} else {
		return err
	}
	return nil
}

func (repo *UserRepository) EndTx(ctx context.Context, err error) error {
	if tx := repo.Tx; tx != nil {
		repo.Tx = nil
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
		return tx.Commit(ctx)
	}
	return nil
}
