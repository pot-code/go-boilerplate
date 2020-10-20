package driver

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

type TxAccessMode int

// transaction access mode
const (
	AccessReadOnly TxAccessMode = iota
	AccessReadWrite
)

type TxDeferrableMode int

// transaction defer mode
const (
	Deferrable TxDeferrableMode = iota
	NotDeferrable
)

// TxOptions Provides a universal option struct across different SQL drivers
type TxOptions struct {
	Isolation      sql.IsolationLevel
	AccessMode     TxAccessMode
	DeferrableMode TxDeferrableMode
}

// ISQLRows Provides a universal query result struct across different SQL drivers
type ISQLRows interface {
	Next() bool
	Scan(dest ...interface{}) (err error)
	Close() error
}

// ITransactionalDB Universal SQL operation interface, to eliminate the gap between different SQL drivers
type ITransactionalDB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (ISQLRows, error)
	BeginTx(ctx context.Context, opts *TxOptions) (ITransactionalDB, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Close(ctx context.Context) error
	Ping() error
}

// DBConfig TODO
type DBConfig struct {
	Driver   string // driver name
	Host     string // server host
	MaxConn  int32  // maximum opening connections number
	Password string // db password
	Port     int    // server port
	Protocol string // connection protocol, eg.tcp
	Query    string // DSN query parameter
	Schema   string // use schema
	User     string // username
}

// SpacePattern check for space, tab or newline
var SpacePattern = regexp.MustCompile(`[\n\t\s]+`)

// DollarPlaceholderPattern check for postgresql style var placeholder
var DollarPlaceholderPattern = regexp.MustCompile(`\$[0-9]+`)

func getDSN(cfg *DBConfig) (DSN string) {
	if cfg.Protocol != "" {
		DSN = fmt.Sprintf("%s:%s@%s(%s:%d)/%s", cfg.User, cfg.Password, cfg.Protocol, cfg.Host, cfg.Port, cfg.Schema)
	} else {
		DSN = fmt.Sprintf("%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Schema)
	}
	if cfg.Query != "" {
		return DSN + "?" + cfg.Query
	}
	return
}

// GetDBConnection create a DB connection from given config
func GetDBConnection(cfg *DBConfig) (conn ITransactionalDB, err error) {
	DSN := getDSN(cfg)
	driver := cfg.Driver

	switch driver {
	case "mysql":
		conn, err = NewMySQLConn(DSN, cfg)
	case "postgres":
		conn, err = NewPostgreSQLConn("postgres://"+DSN, cfg)
	default:
		err = fmt.Errorf("Unsupported driver: %s", driver)
	}
	return
}

func shouldLogError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) {
		return false
	}
	return true
}

func logQueryArgs(args []interface{}) []interface{} {
	logArgs := make([]interface{}, 0, len(args))

	for _, a := range args {
		switch v := a.(type) {
		case []byte:
			if len(v) < 64 {
				a = hex.EncodeToString(v)
			} else {
				a = fmt.Sprintf("%x (truncated %d bytes)", v[:64], len(v)-64)
			}
		case string:
			if len(v) > 64 {
				a = fmt.Sprintf("%s (truncated %d bytes)", v[:64], len(v)-64)
			}
		}
		logArgs = append(logArgs, a)
	}

	return logArgs
}

// func log(logger *zap.Logger,
// 	ctx context.Context,
// 	method, query string,
// 	err error,
// 	fields ...zap.Field,
// ) {
// 	if err != nil {
// 		if shouldLogError(err) {
// 			logger.Error(err.Error(), zap.String("db.sql", query),
// 				zap.String("db.method", method),
// 			)
// 		}
// 	} else {
// 		endTime := time.Now()
// 		logger.Debug("", zap.String("db.sql", query),
// 			zap.Duration("time", endTime.Sub(startTime)),
// 			zap.String("db.method", "Exec"),
// 			zap.Any("args", logQueryArgs(args)))
// 	}
// }
