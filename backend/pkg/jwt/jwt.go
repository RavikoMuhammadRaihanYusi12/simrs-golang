// Package jwt issues and verifies HS256 tokens.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// demoUser is Phase 1 placeholder auth before persistence.
type demoUser struct {
	password string
	role     string
}

var demoUsers = map[string]demoUser{
	"admin":    {password: "admin123", role: "admin"},
	"dokter":   {password: "dokter123", role: "doctor"},
	"loket":    {password: "loket123", role: "nurse"},
	"kasir":    {password: "kasir123", role: "cashier"},
	"apoteker": {password: "apoteker123", role: "pharmacy"},
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (s *JWTService) Login(username, password string) (string, error) {
	u, ok := demoUsers[username]
	if !ok || u.password != password {
		return "", ErrInvalidCredentials
	}
	claims := jwt.MapClaims{
		"sub":  username,
		"role": u.role,
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *JWTService) Verify(token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// RequireAllowlist returns true when claims role is in allow.
func RequireAllowlist(claims jwt.MapClaims, allow ...string) bool {
	role, _ := claims["role"].(string)
	for _, a := range allow {
		if role == a {
			return true
		}
	}
	return false
}
