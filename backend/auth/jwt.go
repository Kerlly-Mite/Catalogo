// Paquete auth: creación y validación de tokens JWT (HS256). No depende de Fiber ni de la BD.
package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Duración del token: el usuario debe volver a iniciar sesión pasadas 2 horas.
const DuracionToken = 2 * time.Hour

// Claims son los datos que viajan dentro del token: sub (id), email, rol, iat y exp.
type Claims struct {
	Email string `json:"email"`
	Rol   string `json:"rol"`
	jwt.RegisteredClaims
}

// GenerarToken firma un JWT HS256 con el secreto recibido.
func GenerarToken(secreto string, usuarioID int, email, rol string) (string, error) {
	ahora := time.Now()
	claims := Claims{
		Email: email,
		Rol:   rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(usuarioID),
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(DuracionToken)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secreto))
}

// ValidarToken verifica firma, algoritmo (solo HS256) y expiración. Devuelve los claims si todo es válido.
func ValidarToken(secreto, token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secreto), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}
	if !parsed.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}
