package user

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/uuid"
)

type UserMySQL struct {
	Conn          driver.ITransactionalDB
	UUIDGenerator uuid.Generator
}

var _ UserRepository = &UserMySQL{}

func NewUserRepository(Conn driver.ITransactionalDB, UUIDGenerator uuid.Generator) *UserMySQL {
	return &UserMySQL{Conn, UUIDGenerator}
}

// FindByCredential query user with provided credential
func (repo *UserMySQL) FindByCredential(ctx context.Context, post *UserModel) (*UserModel, error) {
	conn := repo.Conn
	username := post.Username
	row, err := conn.QueryContext(ctx, `SELECT id, username, password, email, login_retry, last_login
	FROM user WHERE username=? OR email=?`, username, username)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		user := new(UserModel)
		if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.LoginRetry, &user.LastLogin); err != nil {
			return nil, err
		}
		return user, nil
	}
	return nil, nil
}

func (repo *UserMySQL) SaveUser(ctx context.Context, post *UserModel) error {
	conn := repo.Conn
	// generate id
	UUIDGenerator := repo.UUIDGenerator
	if uuid, err := UUIDGenerator.Generate(); err == nil {
		post.ID = uuid
	} else {
		return err
	}

	_, err := conn.ExecContext(ctx, `INSERT INTO user(id, username, password, email, last_login)
	VALUES(?,?,?,?,?)`, post.ID, post.Username, post.Password, post.Email, post.LastLogin)

	if err, ok := err.(*mysql.MySQLError); ok && err.Number == 1062 {
		return ErrDuplicatedUser
	}
	return err
}

func (repo *UserMySQL) UpdateLogin(ctx context.Context, post *UserModel) error {
	conn := repo.Conn
	_, err := conn.ExecContext(ctx, `UPDATE user
	SET login_retry=?,
			last_login=?
	WHERE id = ?;`, post.LoginRetry, post.LastLogin, post.ID)
	return err
}

func (repo *UserMySQL) BeginTx(ctx context.Context) (driver.ITransactionalDB, error) {
	return repo.Conn.BeginTx(ctx, &driver.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
}
