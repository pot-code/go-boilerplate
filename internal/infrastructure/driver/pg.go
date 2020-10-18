package driver

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"go.uber.org/zap"
)

type PGWrapper struct {
	db *pgxpool.Pool
}

type PGWrapperTx struct {
	tx pgx.Tx
}

type PGExecResult struct {
	ct pgconn.CommandTag
}

type PGQueryResult struct {
	rows pgx.Rows
}

// NewPostgreSQLConn Returns a postgreSQL connection pool
func NewPostgreSQLConn(dsn string, cfg *DBConfig) (ITransactionalDB, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// the lib will handle logging
	poolConfig.MaxConns = cfg.MaxConn
	conn, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	return &PGWrapper{conn}, err
}

func (pr PGExecResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (pr PGExecResult) RowsAffected() (int64, error) {
	return pr.ct.RowsAffected(), nil
}

func (pr PGQueryResult) Next() bool {
	return pr.rows.Next()
}
func (pr PGQueryResult) Scan(dest ...interface{}) (err error) {
	return pr.rows.Scan(dest...)
}
func (pr PGQueryResult) Close() error {
	pr.rows.Close()
	return nil
}

func (pw *PGWrapper) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	logger := infra.ExtractLoggerFromContext(ctx)
	startTime := time.Now()

	txConfig := pgTxOptionAdapter(opts)
	tx, err := pw.db.BeginTx(ctx, txConfig)
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
	return &PGWrapperTx{tx}, err
}

func pgTxOptionAdapter(opts *TxOptions) pgx.TxOptions {
	if opts == nil {
		return pgx.TxOptions{}
	}
	iso := pgx.TxIsoLevel(strings.ToLower(opts.Isolation.String()))

	var access pgx.TxAccessMode
	if opts.AccessMode == AccessReadOnly {
		access = pgx.ReadOnly
	} else {
		access = pgx.ReadWrite
	}

	var deferrable pgx.TxDeferrableMode
	if opts.DeferrableMode == Deferrable {
		deferrable = pgx.Deferrable
	} else {
		deferrable = pgx.NotDeferrable
	}
	return pgx.TxOptions{
		IsoLevel:       iso,
		AccessMode:     access,
		DeferrableMode: deferrable,
	}
}

func (pw *PGWrapper) Commit(ctx context.Context) error {
	return nil
}

func (pw *PGWrapper) Rollback(ctx context.Context) error {
	return nil
}

// Close close the whole pool, you better know what you are doing
func (pw *PGWrapper) Close(ctx context.Context) error {
	pw.db.Close()
	return nil
}

func (pw *PGWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logger := infra.ExtractLoggerFromContext(ctx)
	startTime := time.Now()

	query = pgsqlAdapter(query)
	res, err := pw.db.Exec(ctx, query, args...)
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
	return &PGExecResult{res}, err
}

func (pw *PGWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	logger := infra.ExtractLoggerFromContext(ctx)
	startTime := time.Now()

	query = pgsqlAdapter(query)
	rows, err := pw.db.Query(ctx, query, args...)
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
	return &PGQueryResult{rows}, err
}

func (pwt *PGWrapperTx) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	panic("create transaction inside a transaction")
}

func (pwt *PGWrapperTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logger := infra.ExtractLoggerFromContext(ctx)
	startTime := time.Now()

	query = pgsqlAdapter(query)
	res, err := pwt.tx.Exec(ctx, query, args...)
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
	return &PGExecResult{res}, err
}

func (pwt *PGWrapperTx) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	logger := infra.ExtractLoggerFromContext(ctx)
	startTime := time.Now()

	query = pgsqlAdapter(query)
	rows, err := pwt.tx.Query(ctx, query, args...)
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
	return &PGQueryResult{rows}, err
}

func (pwt *PGWrapperTx) Commit(ctx context.Context) error {
	logger := infra.ExtractLoggerFromContext(ctx)
	startTime := time.Now()
	err := pwt.tx.Commit(ctx)
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

func (pwt *PGWrapperTx) Rollback(ctx context.Context) error {
	logger := infra.ExtractLoggerFromContext(ctx)
	startTime := time.Now()
	err := pwt.tx.Rollback(ctx)
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

func (pwt *PGWrapperTx) Close(ctx context.Context) error {
	return nil
}

func pgsqlAdapter(query string) string {
	return SpacePattern.ReplaceAllString(query, " ")
}
