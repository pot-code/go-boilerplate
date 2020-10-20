package driver

import "time"

// KeyValueDB define a key-value storage interface
type KeyValueDB interface {
	SetEX(key string, value string, expiration time.Duration) error
	Get(key string) (string, error)
	Exists(key string) (bool, error)
	Ping() error
}
