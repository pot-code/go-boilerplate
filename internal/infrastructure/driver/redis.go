package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// RedisClient .
type RedisClient struct {
	conn *redis.Client
}

// NewRedisClient create a redis client
func NewRedisClient(host string, port int, password string) *RedisClient {
	conn := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
	})
	return &RedisClient{
		conn: conn,
	}
}

// SetEX implement KeyValueDB
func (rdb *RedisClient) SetEX(key string, value string, expiration time.Duration) error {
	return rdb.conn.Set(ctx, key, value, expiration).Err()
}

// Get implement KeyValueDB
func (rdb *RedisClient) Get(key string) (string, error) {
	cmd := rdb.conn.Get(ctx, key)
	return cmd.Result()
}

// Exists implement KeyValueDB
func (rdb *RedisClient) Exists(key string) (bool, error) {
	cmd := rdb.conn.Exists(ctx, key)
	if ok, err := cmd.Result(); err == nil {
		return ok == 1, nil
	} else {
		return false, err
	}
}
