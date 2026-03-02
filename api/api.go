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

type JobStart struct {
	Total   int
	Success int
	Failed  int
}

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
	dbConn, err := db.GetDatabasePool(ctx)
	if err != nil {
		return err
	}

	err = db.RemoveExpiredAlerts(ctx, dbConn)
	if err != nil {
		return err
	}
	rlog.Info("Removed expired alerts")

	eligibleReminders, err := event.GetEligibleReminders(ctx, dbConn)
	if err != nil {
		return err
	}

	rlog.Info(fmt.Sprintf("Found %d eligible reminders", len(eligibleReminders)))

	if len(eligibleReminders) == 0 {
		return nil
	}

	alerts, err := event.CreateAlertsForEligibleReminders(ctx, dbConn, eligibleReminders)
	if err != nil {
		return err
	}
	rlog.Info(fmt.Sprintf("Created %d alerts", len(alerts)))

	return err
}

//encore:api private
func SendAlertsApi(ctx context.Context) error {
	dbConn, err := db.GetDatabasePool(ctx)
	if err != nil {
		return err
	}

	alerts, err := db.GetAlertsForSending(ctx, dbConn)
	if err != nil {
		return err
	}

	stats := JobStart{Total: len(alerts)}
	rlog.Info(fmt.Sprintf("Found %d alerts to send", stats.Total))
	if stats.Total == 0 {
		return nil
	}

	for _, alert := range alerts {
		_, err := message.SendAlertEmail(alert)
		if err != nil {
			stats.Failed++
			rlog.Error("Error sending alert email", "err", err.Error(), "alertId", alert.ID, "reminderId", alert.ReminderID)
			continue
		}

		stats.Success++
	}
	rlog.Info(fmt.Sprintf("SendAlertsApi summary total=%d success=%d failed=%d", stats.Total, stats.Success, stats.Failed))

	return nil
}

//encore:api private
func RenewRepeatableRemindersApi(ctx context.Context) error {
	dbConn, err := db.GetDatabasePool(ctx)
	if err != nil {
		return err
	}

	reminders, err := db.GetRepeatableReminders(ctx, dbConn)
	if err != nil {
		return err
	}
	stats := JobStart{Total: len(reminders)}
	rlog.Info(fmt.Sprintf("Found %d repeatable reminders", stats.Total))
	if stats.Total == 0 {
		return nil
	}

	for _, reminder := range reminders {
		nextAlertDate := event.GetNextAlertDate(reminder.Repeat, reminder.ReminderDate)
		if nextAlertDate == nil {
			continue
		}

		err = db.UpdateReminderNextAlertDate(ctx, dbConn, reminder.ID, *nextAlertDate)
		if err != nil {
			stats.Failed++
			rlog.Error("Error renewing reminder", "err", err.Error(), "reminderId", reminder.ID, "repeat", string(reminder.Repeat))
			continue
		}

		stats.Success++
	}

	rlog.Info(fmt.Sprintf("RenewRepeatableRemindersApi summary total=%d success=%d failed=%d", stats.Total, stats.Success, stats.Failed))

	return nil
}

type AcknowledgeAlertQueryParams struct {
	Token string `query:"token"`
}

type AcknowledgeAlertResponse struct {
	Message string `json:"message"`
}

//encore:api public method=GET path=/alerts/acknowledge tag:acknowledge
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

	dbConn, err := db.GetDatabasePool(ctx)
	if err != nil {
		return &AcknowledgeAlertResponse{Message: ""}, err
	}

	err = db.AcknowledgeAlert(ctx, dbConn, claims.AlertID)
	if err != nil {
		return &AcknowledgeAlertResponse{Message: ""}, err
	}

	return &AcknowledgeAlertResponse{Message: "Reminder acknowledged"}, nil
}
