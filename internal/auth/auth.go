package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidToken is returned by ParseToken for any malformed, expired,
// or tampered token.
var ErrInvalidToken = errors.New("invalid or expired token")

// jwtSecret is loaded once from the environment. In production this MUST
// be set to a long random value — see .env.example.
func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fallback only for local development so the app doesn't crash
		// if you forget to set it. Never rely on this in a real deployment.
		secret = "dev-only-insecure-secret-change-me"
	}
	return []byte(secret)
}

// HashPassword turns a plaintext password into a bcrypt hash for storage.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a stored bcrypt hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// claims embeds the standard JWT fields plus our own user_id claim.
type claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken issues a signed JWT for the given user, valid for 24 hours.
func GenerateToken(userID int64) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(jwtSecret())
}

// ParseToken validates a token string and returns the embedded user ID.
func ParseToken(tokenString string) (int64, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		// Guard against an attacker submitting a token signed with a
		// different algorithm than the one we use.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return jwtSecret(), nil
	})
	if err != nil || !parsed.Valid {
		return 0, ErrInvalidToken
	}
	c, ok := parsed.Claims.(*claims)
	if !ok {
		return 0, ErrInvalidToken
	}
	return c.UserID, nil
}
