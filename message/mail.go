package message

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"encore.dev"
	"github.com/6tail/lunar-go/calendar"
	"github.com/hazymoon22/lunar-worker/acktoken"
	"github.com/hazymoon22/lunar-worker/db"
	"github.com/mailgun/mailgun-go/v4"
)

var secrets struct {
	MailgunApiKey  string
	MailgunSandBox string
}

func SendAlertEmail(alert db.Alert) (string, error) {
	// Get base URL and convert to string
	baseUrl := encore.Meta().APIBaseURL.String()

	acknowledgeToken, err := acktoken.Generate(alert.ID)
	if err != nil {
		return "", err
	}

	// URL encode the token to handle special characters
	encodedToken := url.QueryEscape(acknowledgeToken)
	acknowledgeReminderEventApiUrl := fmt.Sprintf("%s/alerts/acknowledge?token=%s", baseUrl, encodedToken)

	now := time.Now().UTC()
	lunarToday := calendar.NewLunarFromDate(now)
	mg := mailgun.NewMailgun(secrets.MailgunSandBox, secrets.MailgunApiKey)
	message := mailgun.NewMessage(
		fmt.Sprintf("Lunar Reminder <postmaster@%s>", secrets.MailgunSandBox),
		alert.Reminder.MailSubject,
		"Click the link to acknowledge this reminder",
		fmt.Sprintf("%s <%s>", alert.Reminder.User.Name, alert.Reminder.User.Email),
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
