package api

import (
	"context"
	"fmt"

	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/rlog"
	"github.com/hazymoon22/lunar-worker/db"
	"github.com/hazymoon22/lunar-worker/event"
	"github.com/hazymoon22/lunar-worker/message"
)

var _ = cron.NewJob("manage-reminders", cron.JobConfig{
	Title:    "Manage alert for lunar events",
	Every:    24 * cron.Hour,
	Endpoint: RenewRepeatableRemindersApi,
})

var _ = cron.NewJob("manage-alerts", cron.JobConfig{
	Title:    "Manage alert for lunar events",
	Every:    24 * cron.Hour,
	Endpoint: ManageAlertsApi,
})

var _ = cron.NewJob("send-alerts", cron.JobConfig{
	Title:    "Send alerts",
	Every:    2 * cron.Hour,
	Endpoint: SendAlertsApi,
})

//encore:api private
func ManageAlertsApi(ctx context.Context) error {
	err := db.RemoveExpiredAlerts(ctx)
	if err != nil {
		return err
	}
	rlog.Info("Removed expired alerts")

	eligibleReminders, err := event.GetEligibleReminders(ctx)
	if err != nil {
		return err
	}

	rlog.Info(fmt.Sprintf("Found %d eligible reminders", len(eligibleReminders)))

	if len(eligibleReminders) == 0 {
		return nil
	}

	alerts, err := event.CreateAlertsForEligibleReminders(ctx, eligibleReminders)
	if err != nil {
		return err
	}
	rlog.Info(fmt.Sprintf("Created %d alerts", len(alerts)))

	return err
}

//encore:api private
func SendAlertsApi(ctx context.Context) error {
	alerts, err := db.GetAlertsForSending(ctx)
	if err != nil {
		return err
	}

	rlog.Info(fmt.Sprintf("Found %d alerts to send", len(alerts)))
	if len(alerts) == 0 {
		return nil
	}

	result := make([]string, 0)
	for _, alert := range alerts {
		id, err := message.SendAlertEmail(alert)
		if err != nil {
			rlog.Error("Error sending alert email", "err", err.Error(), "alert", alert)
			continue
		}

		result = append(result, id)
	}
	rlog.Info(fmt.Sprintf("Sent %d alert emails", len(result)))

	return err
}

//encore:api private
func RenewRepeatableRemindersApi(ctx context.Context) error {
	reminders, err := db.GetRepeatableReminders(ctx)
	if err != nil {
		return err
	}
	rlog.Info(fmt.Sprintf("Found %d repeatable reminders", len(reminders)))
	if len(reminders) == 0 {
		return nil
	}

	numRenewReminders := 0
	for _, reminder := range reminders {
		nextAlertDate := event.GetNextAlertDate(reminder.Repeat, reminder.ReminderDate)
		if nextAlertDate == nil {
			continue
		}

		err = db.UpdateReminderNextAlertDate(ctx, reminder.ID, *nextAlertDate)
		if err != nil {
			continue
		}

		numRenewReminders++
	}

	rlog.Info(fmt.Sprintf("Renewed %d reminders", numRenewReminders))

	return err
}

type AcknowledgeAlertQueryParams struct {
	Token string `query:"token"`
}

type AcknowledgeAlertResponse struct {
	Message string `json:"message"`
}

//encore:api public method=GET path=/alerts/acknowledge
func AcknowledgeAlertApi(ctx context.Context, params *AcknowledgeAlertQueryParams) (*AcknowledgeAlertResponse, error) {
	if params == nil {
		return &AcknowledgeAlertResponse{Message: ""}, errs.B().
			Code(errs.Unauthenticated).
			Err()
	}

	claims, err := message.VerifyAcknowledgeAlertToken(params.Token)
	if err != nil {
		return &AcknowledgeAlertResponse{Message: ""}, err
	}

	err = db.AcknowledgeAlert(ctx, claims.AlertID)
	if err != nil {
		return &AcknowledgeAlertResponse{Message: ""}, err
	}

	return &AcknowledgeAlertResponse{Message: "Reminder acknowledged"}, nil
}
