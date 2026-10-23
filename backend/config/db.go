// Paquete config: configuración de la aplicación (variables de entorno, PostgreSQL y Redis).
package config

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// LoadEnv carga las variables del archivo .env (si existe) al entorno del proceso.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env: se usarán variables del sistema o valores por defecto")
	}
}

// ConnectDB crea el pool de conexiones a PostgreSQL y comprueba que responde.
// (Fase 4) Ya NO guarda el pool en una variable global: lo devuelve para inyectarlo donde se necesite.
func ConnectDB() (*pgxpool.Pool, error) {
	// En la nube (Render) llega una sola URL completa en DATABASE_URL; en local se arma con las variables DB_*.
	dsn := GetEnv("DATABASE_URL", "")
	if dsn == "" {
		host := GetEnv("DB_HOST", "localhost")
		port := GetEnv("DB_PORT", "5432")
		user := GetEnv("DB_USER", "catalogo_user")
		pass := GetEnv("DB_PASSWORD", "catalogo_pass")
		name := GetEnv("DB_NAME", "multicatalogo")
		// DB_SSLMODE: "disable" en local (Docker); "require" si la base exige TLS.
		ssl := GetEnv("DB_SSLMODE", "disable")

		// QueryEscape evita problemas si la contraseña tiene caracteres especiales (@, :, /...).
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.QueryEscape(user), url.QueryEscape(pass), host, port, name, url.QueryEscape(ssl))
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("configuración de BD inválida: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("PostgreSQL no responde (¿está levantado Docker?): %w", err)
	}

	// Se registra solo host y base (nunca la contraseña).
	log.Printf("Conectado a PostgreSQL (%s/%s)", cfg.ConnConfig.Host, cfg.ConnConfig.Database)
	return pool, nil
}
