package driver

import (
	"context"
	"database/sql"
	"strings"
	"time"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"go.uber.org/zap"
)

// SQLWrapper Wraps a *sql.db object and provides the implementation of ITransactionalDB.
//
// it uses zap for default logging
type SQLWrapper struct {
	db *sql.DB
}

// SQLWrapperTx transaction wrapper
type SQLWrapperTx struct {
	tx *sql.Tx
}

// NewMySQLConn Returns a MySQL connection pool
func NewMySQLConn(dsn string, cfg *DBConfig) (ITransactionalDB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(int(cfg.MaxConn))
	return &SQLWrapper{conn}, err
}

// BeginTx start a new transaction context
func (mw *SQLWrapper) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	logger := ctx.Value(infra.ContextLoggerKey).(*zap.Logger)
	startTime := time.Now()

	txConfig := mysqlTxOptionAdapter(opts)
	tx, err := mw.db.BeginTx(ctx, txConfig)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(err.Error(), zap.String("db.method", "BeginTx"))
		}
	} else {
		endTime := time.Now()
		logger.Debug("", zap.Duration("db.time", endTime.Sub(startTime)),
			zap.String("db.method", "BeginTx"),
		)
	}
	return &SQLWrapperTx{tx}, err
}

func mysqlTxOptionAdapter(opts *TxOptions) *sql.TxOptions {
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
	logger := ctx.Value(infra.ContextLoggerKey).(*zap.Logger)
	startTime := time.Now()

	query = mysqlAdapter(query)
	res, err := mw.db.ExecContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(err.Error(), zap.String("db.sql", query),
				zap.String("db.method", "Exec"),
				zap.Any("db.args", logQueryArgs(args)))
		}
	} else {
		endTime := time.Now()
		logger.Debug("", zap.String("db.sql", query),
			zap.Duration("db.time", endTime.Sub(startTime)),
			zap.String("db.method", "Exec"),
			zap.Any("db.args", logQueryArgs(args)))
	}
	return res, err
}

func (mw *SQLWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	logger := ctx.Value(infra.ContextLoggerKey).(*zap.Logger)
	startTime := time.Now()

	query = mysqlAdapter(query)
	rows, err := mw.db.QueryContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(err.Error(), zap.String("db.sql", query),
				zap.String("db.method", "Query"),
				zap.Any("db.args", logQueryArgs(args)))
		}
	} else {
		endTime := time.Now()
		logger.Debug("", zap.String("db.sql", query),
			zap.Duration("db.time", endTime.Sub(startTime)),
			zap.String("db.method", "Query"),
			zap.Any("db.args", logQueryArgs(args)))
	}
	return rows, err
}

func (mwt *SQLWrapperTx) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	panic("create transaction inside a transaction")
}

func (mwt *SQLWrapperTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logger := ctx.Value(infra.ContextLoggerKey).(*zap.Logger)
	startTime := time.Now()

	query = mysqlAdapter(query)
	res, err := mwt.tx.ExecContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(err.Error(), zap.String("db.sql", query),
				zap.String("db.method", "Exec"),
				zap.Any("db.args", logQueryArgs(args)))
		}
	} else {
		endTime := time.Now()
		logger.Debug("", zap.String("db.sql", query),
			zap.Duration("db.time", endTime.Sub(startTime)),
			zap.String("db.method", "Exec"),
			zap.Any("db.args", logQueryArgs(args)))
	}
	return res, err
}

func (mwt *SQLWrapperTx) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	logger := ctx.Value(infra.ContextLoggerKey).(*zap.Logger)
	startTime := time.Now()

	query = mysqlAdapter(query)
	rows, err := mwt.tx.QueryContext(ctx, query, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(err.Error(), zap.String("db.sql", query),
				zap.String("db.method", "Query"),
				zap.Any("db.args", logQueryArgs(args)))
		}
	} else {
		endTime := time.Now()
		logger.Debug("", zap.String("db.sql", query),
			zap.Duration("db.time", endTime.Sub(startTime)),
			zap.String("db.method", "Query"),
			zap.Any("db.args", logQueryArgs(args)))
	}
	return rows, err
}

func (mwt *SQLWrapperTx) Commit(ctx context.Context) error {
	logger := ctx.Value(infra.ContextLoggerKey).(*zap.Logger)
	startTime := time.Now()
	err := mwt.tx.Commit()
	if err != nil {
		if shouldLogError(err) {
			logger.Error(err.Error(), zap.String("db.method", "Commit"))
		}
	} else {
		endTime := time.Now()
		logger.Debug("", zap.Duration("db.time", endTime.Sub(startTime)),
			zap.String("db.method", "Commit"),
		)
	}
	return err
}

func (mwt *SQLWrapperTx) Rollback(ctx context.Context) error {
	logger := ctx.Value(infra.ContextLoggerKey).(*zap.Logger)
	startTime := time.Now()
	err := mwt.tx.Rollback()
	if err != nil {
		if shouldLogError(err) {
			logger.Error(err.Error(), zap.String("db.method", "RollBack"))
		}
	} else {
		endTime := time.Now()
		logger.Debug("", zap.Duration("db.time", endTime.Sub(startTime)),
			zap.String("db.method", "RollBack"),
		)
	}
	return err
}

func (mwt *SQLWrapperTx) Close(ctx context.Context) error {
	return nil
}

func mysqlAdapter(query string) string {
	query = strings.Replace(query, "\"", "`", -1)
	query = DollarPlaceholderPattern.ReplaceAllString(query, "?")
	query = SpacePattern.ReplaceAllString(query, " ")
	return query
}
