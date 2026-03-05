package acktoken

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerify(t *testing.T) {
	secrets.JwtSecret = "test-secret"

	token, err := Generate("alert-123")
	require.NoError(t, err)

	claims, err := Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "alert-123", claims.AlertID)
}

func TestVerifyInvalidSignature(t *testing.T) {
	secrets.JwtSecret = "secret-a"
	token, err := Generate("alert-123")
	require.NoError(t, err)

	secrets.JwtSecret = "secret-b"
	_, err = Verify(token)
	require.Error(t, err)
}

func TestVerifyExpired(t *testing.T) {
	secrets.JwtSecret = "test-secret"

	claims := Claims{
		AlertID: "alert-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
			Issuer:    "lunar-reminder",
			Subject:   "reminder:alert-123",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secrets.JwtSecret))
	require.NoError(t, err)

	_, err = Verify(tokenString)
	require.Error(t, err)
}
