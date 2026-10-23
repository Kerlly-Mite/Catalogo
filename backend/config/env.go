package config

import (
	"os"
	"strconv"
	"strings"
)

// GetEnv devuelve el valor de una variable de entorno o un valor por defecto si está vacía.
func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// GetEnvBool interpreta una variable de entorno como booleano ("true", "1", "si").
func GetEnvBool(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "si" || v == "sí"
}

// GetEnvInt interpreta una variable de entorno como entero.
func GetEnvInt(key string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil {
		return n
	}
	return def
}
