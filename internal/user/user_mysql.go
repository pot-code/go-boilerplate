package user

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/pot-code/go-boilerplate/internal/domain"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/uuid"
)

type UserRepository struct {
	Conn          driver.ITransactionalDB
	UUIDGenerator uuid.UUIDGenerator
}

func NewUserRepository(Conn driver.ITransactionalDB, UUIDGenerator uuid.UUIDGenerator) *UserRepository {
	return &UserRepository{Conn, UUIDGenerator}
}

// FindByCredential query user with provided credential
func (repo *UserRepository) FindByCredential(ctx context.Context, post *domain.UserModel) (*domain.UserModel, error) {
	conn := repo.Conn
	username := post.Username
	row, err := conn.QueryContext(ctx, `SELECT id, username, password, email, login_retry, last_login
	FROM user WHERE username=? OR email=?`, username, username)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		user := new(domain.UserModel)
		if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.LoginRetry, &user.LastLogin); err != nil {
			return nil, err
		}
		return user, nil
	}
	return nil, nil
}

func (repo *UserRepository) SaveUser(ctx context.Context, post *domain.UserModel) error {
	conn := repo.Conn
	// generate id
	UUIDGenerator := repo.UUIDGenerator
	if uuid, err := UUIDGenerator.Generate(); err == nil {
		post.ID = uuid
	} else {
		return err
	}

	_, err := conn.ExecContext(ctx, `INSERT INTO user(id, username, password, email)
	VALUES(?,?,?,?)`, post.ID, post.Username, post.Password, post.Email)

	if err, ok := err.(*mysql.MySQLError); ok && err.Number == 1062 {
		return domain.ErrDuplicatedUser
	}
	return err
}

func (repo *UserRepository) UpdateUser(ctx context.Context, post *domain.UserModel) error {
	conn := repo.Conn
	_, err := conn.ExecContext(ctx, `UPDATE user
	SET email=?,
			login_retry=?,
			last_login=?
	WHERE id = ?;`, post.Email, post.LoginRetry, post.LastLogin, post.ID)
	return err
}

func (repo *UserRepository) BeginTx(ctx context.Context) (driver.ITransactionalDB, error) {
	return repo.Conn.BeginTx(ctx, &driver.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
}
