package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis crea el cliente de Redis y comprueba que responde.
// Lee REDIS_URL (URI completa) o, si no existe, REDIS_HOST / REDIS_PORT / REDIS_PASSWORD.
func ConnectRedis() (*redis.Client, error) {
	var opts *redis.Options

	if uri := GetEnv("REDIS_URL", ""); uri != "" {
		parsed, err := redis.ParseURL(uri)
		if err != nil {
			return nil, fmt.Errorf("REDIS_URL inválida: %w", err)
		}
		opts = parsed
	} else {
		opts = &redis.Options{
			Addr:     GetEnv("REDIS_HOST", "localhost") + ":" + GetEnv("REDIS_PORT", "6379"),
			Password: GetEnv("REDIS_PASSWORD", ""),
			DB:       0,
		}
	}
	// Tiempos cortos: si Redis se cae, la API no debe quedarse esperando.
	opts.DialTimeout = 2 * time.Second
	opts.ReadTimeout = 2 * time.Second
	opts.WriteTimeout = 2 * time.Second

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("Redis no responde (¿está levantado Docker?): %w", err)
	}

	log.Printf("Conectado a Redis (%s)", opts.Addr)
	return client, nil
}
