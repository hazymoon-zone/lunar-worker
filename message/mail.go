package message

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"encore.dev"
	"encore.dev/rlog"
	"github.com/6tail/lunar-go/calendar"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hazymoon22/lunar-worker/db"
	"github.com/mailgun/mailgun-go/v4"
)

var secrets struct {
	MailgunApiKey  string
	MailgunSandBox string
	JwtSecret      string
}

type ReminderClaims struct {
	AlertID string `json:"reminder_id"`
	jwt.RegisteredClaims
}

func generateAcknowledgeAlertToken(alertId string) (string, error) {
	// Create claims with expiration
	claims := ReminderClaims{
		AlertID: alertId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "lunar-reminder",
			Subject:   fmt.Sprintf("reminder:%s", alertId),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(secrets.JwtSecret))
	if err != nil {
		rlog.Error("Failed to sign token", "err", err.Error())
		return "", err
	}

	return tokenString, nil
}

func VerifyAcknowledgeAlertToken(tokenString string) (*ReminderClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &ReminderClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secrets.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract and validate claims
	claims, ok := token.Claims.(*ReminderClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token")
	}

	return claims, nil
}

func SendAlertEmail(alert db.Alert) (string, error) {
	// Get base URL and convert to string
	baseUrl := encore.Meta().APIBaseURL.String()

	acknowledgeToken, err := generateAcknowledgeAlertToken(alert.ID)
	if err != nil {
		rlog.Error("Error generating acknowledge token", "err", err.Error(), "alertId", alert.ID)
		return "", err
	}

	// URL encode the token to handle special characters
	encodedToken := url.QueryEscape(acknowledgeToken)
	acknowledgeReminderEventApiUrl := fmt.Sprintf("%s/alerts/acknowledge?token=%s", baseUrl, encodedToken)

	now := time.Now()
	lunarToday := calendar.NewLunarFromDate(now)
	mg := mailgun.NewMailgun(secrets.MailgunSandBox, secrets.MailgunApiKey)
	message := mailgun.NewMessage(
		fmt.Sprintf("Lunar Reminder <postmaster@%s>", secrets.MailgunSandBox),
		alert.Reminder.MailSubject,
		"Click the link to acknowledge this reminder",
		"Bùi Đức Huy <huybui150396@gmail.com>",
	)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
		<html>
		<head><meta charset="UTF-8"></head>
		<body>
		%s
		<p>Today is %02d/%02d Lunar.</p>
		<p><a href="%s">Acknowledge Reminder</a></p>
		</body>
		</html>`, alert.Reminder.MailBody, lunarToday.GetDay(), lunarToday.GetMonth(), acknowledgeReminderEventApiUrl)
	message.SetHTML(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, message)

	return id, err
}
