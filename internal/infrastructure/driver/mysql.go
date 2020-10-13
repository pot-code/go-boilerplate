package driver

import (
	"context"
	"database/sql"
	"strings"
	"time"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// NewMySQLConn Returns a MySQL connection pool
func NewMySQLConn(dsn string, cfg *DBConfig) (ITransactionalDB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	logger := cfg.Logger.With(zap.String("driver", cfg.Driver), zap.String("database", cfg.Schema))
	logger.Debug("Create mysql connection instance", zap.Any("config", cfg))
	conn.SetMaxOpenConns(int(cfg.MaxConn))
	return &SQLWrapper{conn, logger}, err
}

// SQLWrapper Wraps a *sql.db object and provides the implementation of ITransactionalDB.
//
// it uses zap for default logging
type SQLWrapper struct {
	db     *sql.DB
	logger *zap.Logger
}

// BeginTx start a new transaction context
func (mw *SQLWrapper) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	logger := mw.logger
	startTime := time.Now()

	txConfig := sqlTxOptionAdapter(opts)
	tx, err := mw.db.BeginTx(ctx, txConfig)
	if err != nil {
		if shouldLogError(err) {
			logger.Error("Exec", zap.String("sql", "begin"), zap.NamedError("err", err))
		}
	} else {
		endTime := time.Now()
		logger.Debug("Exec", zap.String("sql", "begin"), zap.Duration("time", endTime.Sub(startTime)))
	}
	return &SQLWrapperTx{tx, logger}, err
}

func sqlTxOptionAdapter(opts *TxOptions) *sql.TxOptions {
	if opts == nil {
		return nil
	}
	iso := opts.Isolation
	readOnly := opts.AccessMode == AccessReadOnly
	return &sql.TxOptions{
		Isolation: iso,
		ReadOnly:  readOnly,
	}
}

func (mw *SQLWrapper) Commit(ctx context.Context) error {
	return nil
}

func (mw *SQLWrapper) Rollback(ctx context.Context) error {
	return nil
}

func (mw *SQLWrapper) Close(ctx context.Context) error {
	return mw.db.Close()
}

func (mw *SQLWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logger := mw.logger
	startTime := time.Now()

	query = sqlAdapter(query)
	res, err := mw.db.ExecContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error("Exec", zap.String("sql", query), zap.Error(err), zap.Any("args", args))
		}
	} else {
		endTime := time.Now()
		logger.Debug("Exec", zap.String("sql", query), zap.Duration("time", endTime.Sub(startTime)), zap.Any("args", args))
	}
	return res, err
}

func (mw *SQLWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	logger := mw.logger
	startTime := time.Now()

	query = sqlAdapter(query)
	rows, err := mw.db.QueryContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error("Query", zap.String("sql", query), zap.Error(err), zap.Any("args", args))
		}
	} else {
		endTime := time.Now()
		logger.Debug("Query", zap.String("sql", query), zap.Duration("time", endTime.Sub(startTime)), zap.Any("args", args))
	}
	return rows, err
}

type SQLWrapperTx struct {
	tx     *sql.Tx
	logger *zap.Logger
}

func (mwt *SQLWrapperTx) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	panic("create transaction inside a transaction")
}

func (mwt *SQLWrapperTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logger := mwt.logger
	startTime := time.Now()

	query = sqlAdapter(query)
	res, err := mwt.tx.ExecContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error("Exec", zap.String("sql", query), zap.NamedError("err", err), zap.Any("args", args))
		}
	} else {
		endTime := time.Now()
		logger.Debug("Exec", zap.String("sql", query), zap.Duration("time", endTime.Sub(startTime)), zap.Any("args", args))
	}
	return res, err
}

func (mwt *SQLWrapperTx) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	logger := mwt.logger
	startTime := time.Now()

	query = sqlAdapter(query)
	rows, err := mwt.tx.QueryContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error("Query", zap.String("sql", query), zap.NamedError("err", err), zap.Any("args", args))
		}
	} else {
		endTime := time.Now()
		logger.Debug("Query", zap.String("sql", query), zap.Duration("time", endTime.Sub(startTime)), zap.Any("args", args))
	}
	return rows, err
}

func (mwt *SQLWrapperTx) Commit(ctx context.Context) error {
	logger := mwt.logger
	startTime := time.Now()
	err := mwt.tx.Commit()
	if err != nil {
		if shouldLogError(err) {
			logger.Error("Exec", zap.String("sql", "commit"), zap.NamedError("err", err))
		}
	} else {
		endTime := time.Now()
		logger.Debug("Exec", zap.String("sql", "commit"), zap.Duration("time", endTime.Sub(startTime)))
	}
	return err
}

func (mwt *SQLWrapperTx) Rollback(ctx context.Context) error {
	logger := mwt.logger
	startTime := time.Now()
	err := mwt.tx.Rollback()
	if err != nil {
		if shouldLogError(err) {
			logger.Error("Exec", zap.String("sql", "rollback"), zap.NamedError("err", err))
		}
	} else {
		endTime := time.Now()
		logger.Debug("Exec", zap.String("sql", "rollback"), zap.Duration("time", endTime.Sub(startTime)))
	}
	return err
}

func (mwt *SQLWrapperTx) Close(ctx context.Context) error {
	return nil
}

func sqlAdapter(query string) string {
	query = strings.Replace(query, "\"", "`", -1)
	query = DollarPlaceholderPattern.ReplaceAllString(query, "?")
	query = SpacePattern.ReplaceAllString(query, " ")
	return query
}
