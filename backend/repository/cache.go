package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implementa CacheRepository con Redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache recibe el cliente ya creado (inyección).
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil // Cache Miss: la clave no existe
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil // Cache Hit
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// DeleteByPrefix recorre las claves con SCAN (no bloquea Redis como KEYS) y las borra.
func (c *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

// NoopCache es una caché "vacía": nunca guarda nada. Se usa si Redis no está disponible al arrancar,
// así la API sigue funcionando (más lenta, pero sin caerse).
type NoopCache struct{}

func (NoopCache) Get(context.Context, string) ([]byte, bool, error)        { return nil, false, nil }
func (NoopCache) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (NoopCache) Delete(context.Context, ...string) error                  { return nil }
func (NoopCache) DeleteByPrefix(context.Context, string) error             { return nil }

// Ping de NoopCache siempre falla: significa que no hay Redis conectado ("degraded" en /api/health).
func (NoopCache) Ping(context.Context) error { return errors.New("sin caché: Redis no está conectado") }
