package message

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyAcknowledgeAlertToken(t *testing.T) {
	secrets.JwtSecret = "test-secret"

	token, err := generateAcknowledgeAlertToken("alert-123")
	require.NoError(t, err)

	claims, err := VerifyAcknowledgeAlertToken(token)
	require.NoError(t, err)
	assert.Equal(t, "alert-123", claims.AlertID)
}

func TestVerifyAcknowledgeAlertTokenInvalidSignature(t *testing.T) {
	secrets.JwtSecret = "secret-a"
	token, err := generateAcknowledgeAlertToken("alert-123")
	require.NoError(t, err)

	secrets.JwtSecret = "secret-b"
	_, err = VerifyAcknowledgeAlertToken(token)
	require.Error(t, err)
}

func TestVerifyAcknowledgeAlertTokenExpired(t *testing.T) {
	secrets.JwtSecret = "test-secret"

	claims := ReminderClaims{
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

	_, err = VerifyAcknowledgeAlertToken(tokenString)
	require.Error(t, err)
}
