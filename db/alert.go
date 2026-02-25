package db

import (
	"context"
	"time"

	"encore.dev/rlog"
	"encore.dev/types/uuid"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
)

type InsertAlertParams struct {
	ReminderID string    `db:"reminder_id"`
	AlertDate  time.Time `db:"alert_date"`
}

type Alert struct {
	ID         string   `db:"id"`
	ReminderID string   `db:"reminder_id"`
	Reminder   Reminder `db:"reminder"`
}

func GetAlertsForToday(ctx context.Context) ([]Alert, error) {
	today := time.Now().Format("2006-01-02")
	isToday := psql.Quote("alert_date").EQ(psql.Arg(today))

	query, args := psql.Select(
		sm.Columns(
			"id",
			"reminder_id",
		),
		sm.From("alert"),
		sm.Where(isToday),
	).MustBuild(ctx)

	db, err := getDatabasePool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		rlog.Error("Error querying alerts from db", "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	alerts := make([]Alert, 0)
	for rows.Next() {
		var row Alert
		if err := rows.Scan(&row.ID, &row.ReminderID); err != nil {
			rlog.Error("Error scanning alerts rows to array", "err", err.Error())
			return nil, err
		}
		alerts = append(alerts, row)
	}

	return alerts, nil
}

func GetAlertsForSending(ctx context.Context) ([]Alert, error) {
	today := time.Now().Format("2006-01-02")
	isToday := psql.Quote("alert_date").EQ(psql.Arg(today))
	isNotAcknowledged := psql.Quote("acknowledged").EQ(psql.Arg(false))

	query, args := psql.Select(
		sm.Columns(
			"a.id",
			"a.reminder_id",
			"r.mail_subject",
			"r.mail_body",
		),
		sm.From("alert").As("a"),
		sm.InnerJoin("reminder").As("r").On(
			psql.Quote("r", "id").EQ(psql.Quote("a", "reminder_id")),
		),
		sm.Where(psql.And(isToday, isNotAcknowledged)),
	).MustBuild(ctx)

	db, err := getDatabasePool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		rlog.Error("Error querying alerts from db", "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	alerts := make([]Alert, 0)
	for rows.Next() {
		var alert Alert
		var reminder Reminder
		if err := rows.Scan(&alert.ID, &alert.ReminderID, &reminder.MailSubject, &reminder.MailBody); err != nil {
			rlog.Error("Error scanning alerts rows to array", "err", err.Error())
			return nil, err
		}
		alert.Reminder = reminder
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func InsertAlert(ctx context.Context, alert InsertAlertParams) (*Alert, error) {
	db, err := getDatabasePool(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		rlog.Error("Error generating uuid", "err", err.Error())
		return nil, err
	}

	query, args := psql.Insert(
		im.Into("alert", "id", "reminder_id", "alert_date"),
		im.Values(psql.Arg(id, alert.ReminderID, alert.AlertDate)),
		im.Returning("id", "reminder_id"),
	).MustBuild(ctx)

	var insertedAlert Alert
	err = db.QueryRow(ctx, query, args...).Scan(&insertedAlert.ID, &insertedAlert.ReminderID)
	if err != nil {
		rlog.Error("Error scanning inserted alert to variable", "err", err.Error(), "alert", alert)
		return nil, err
	}

	return &insertedAlert, nil
}

func RemoveExpiredAlerts(ctx context.Context) error {
	today := time.Now().Format("2006-01-02")
	beforeToday := psql.Quote("alert_date").LT(psql.Arg(today))

	query, args := psql.Delete(
		dm.From("alert"),
		dm.Where(beforeToday),
	).MustBuild(ctx)

	db, err := getDatabasePool(ctx)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, query, args...)
	if err != nil {
		rlog.Error("Error removing expired alerts from db", "err", err.Error())
		return err
	}

	return nil
}

func AcknowledgeAlert(ctx context.Context, alertId string) error {
	query, args := psql.Update(
		um.Table("alert"),
		um.SetCol("acknowledged").To(psql.Arg(true)),
		um.Where(psql.Quote("id").EQ(psql.Arg(alertId))),
	).MustBuild(ctx)

	db, err := getDatabasePool(ctx)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, query, args...)
	if err != nil {
		rlog.Error("Error updating acknowledged status for alert", "err", err.Error(), "alertId", alertId)
		return err
	}

	return err
}
