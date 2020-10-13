package driver

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// NewPostgreSQLConn Returns a postgreSQL connection pool
func NewPostgreSQLConn(dsn string, cfg *DBConfig) (ITransactionalDB, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	logger := cfg.Logger.With(zap.String("driver", cfg.Driver), zap.String("database", cfg.Schema))
	logger.Debug("Create pgsql connection instance", zap.Any("config", cfg))
	// the lib will handle logging
	poolConfig.ConnConfig.Logger = zapadapter.NewLogger(logger)
	poolConfig.MaxConns = cfg.MaxConn
	conn, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	return &PGWrapper{conn}, err
}

type PGExecResult struct {
	ct pgconn.CommandTag
}

func (pr PGExecResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (pr PGExecResult) RowsAffected() (int64, error) {
	return pr.ct.RowsAffected(), nil
}

type PGQueryResult struct {
	rows pgx.Rows
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

type PGWrapper struct {
	db *pgxpool.Pool
}

func (pw *PGWrapper) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	txConfig := pgTxOptionAdapter(opts)
	tx, err := pw.db.BeginTx(ctx, txConfig)
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
	query = pgsqlAdapter(query)
	res, err := pw.db.Exec(ctx, query, args...)
	return &PGExecResult{res}, err
}

func (pw *PGWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	query = pgsqlAdapter(query)
	rows, err := pw.db.Query(ctx, query, args...)
	return &PGQueryResult{rows}, err
}

type PGWrapperTx struct {
	tx pgx.Tx
}

func (pwt *PGWrapperTx) BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error) {
	panic("create transaction inside a transaction")
}

func (pwt *PGWrapperTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query = pgsqlAdapter(query)
	res, err := pwt.tx.Exec(ctx, query, args...)
	return &PGExecResult{res}, err
}

func (pwt *PGWrapperTx) QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error) {
	query = pgsqlAdapter(query)
	rows, err := pwt.tx.Query(ctx, query, args...)
	return &PGQueryResult{rows}, err
}

func (pwt *PGWrapperTx) Commit(ctx context.Context) error {
	return pwt.tx.Commit(ctx)
}

func (pwt *PGWrapperTx) Rollback(ctx context.Context) error {
	return pwt.tx.Rollback(ctx)
}

func (pwt *PGWrapperTx) Close(ctx context.Context) error {
	return nil
}

func pgsqlAdapter(query string) string {
	return SpacePattern.ReplaceAllString(query, " ")
}
