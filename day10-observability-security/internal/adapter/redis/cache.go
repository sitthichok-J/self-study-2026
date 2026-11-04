// Package redis implements port.Cache (works with Redis or Valkey).
package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Cache struct{ c *goredis.Client }

func New(addr string) *Cache {
	return &Cache{c: goredis.NewClient(&goredis.Options{
		Addr: addr, DialTimeout: time.Second, ReadTimeout: 500 * time.Millisecond, WriteTimeout: 500 * time.Millisecond,
	})}
}

func (c *Cache) Ping(ctx context.Context) error { return c.c.Ping(ctx).Err() }
func (c *Cache) Close() error                   { return c.c.Close() }

func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	b, err := c.c.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (c *Cache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.c.Set(ctx, key, val, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	return c.c.Del(ctx, keys...).Err()
}
