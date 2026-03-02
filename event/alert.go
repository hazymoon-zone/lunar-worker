package event

import (
	"context"
	"fmt"
	"time"

	"encore.dev/rlog"
	"github.com/hazymoon22/lunar-worker/db"
)

func CreateAlertsForEligibleReminders(ctx context.Context, dbConn db.DBTX, eligibleReminders []db.Reminder) ([]db.Alert, error) {
	alerts, err := db.GetAlertsForToday(ctx, dbConn)
	if err != nil {
		return nil, err
	}
	rlog.Info(fmt.Sprintf("Found %d alerts for today", len(alerts)))

	now := time.Now().UTC()
	result := make([]db.Alert, 0)
	for _, reminder := range eligibleReminders {
		hasAlertAlready := false
		for _, alert := range alerts {
			if alert.ReminderID == reminder.ID {
				hasAlertAlready = true
				break
			}
		}

		if hasAlertAlready {
			continue
		}

		insertAlert := db.InsertAlertParams{
			ReminderID: reminder.ID,
			AlertDate:  time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		}

		alert, err := db.InsertAlert(ctx, dbConn, insertAlert)
		if err != nil {
			continue
		}
		result = append(result, *alert)
	}

	return result, nil
}
