package db

import (
	"context"
	"time"

	"encore.dev/rlog"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
)

type RepeatMode string

const (
	RepeatModeMonthly RepeatMode = "monthly"
	RepeatModeYearly  RepeatMode = "yearly"
)

type Reminder struct {
	ID            string
	ReminderDate  time.Time
	NextAlertDate time.Time
	Repeat        RepeatMode
	AlertBefore   *int32
	MailSubject   string
	MailBody      string
	User          User
}

func GetRemindersFromToday(ctx context.Context, db DBTX) ([]Reminder, error) {
	today := time.Now().UTC().Format("2006-01-02")
	afterOrEqualToday := psql.Quote("next_alert_date").GTE(psql.Arg(today))

	query, args := psql.Select(
		sm.Columns("id", "next_alert_date", "alert_before"),
		sm.From("reminder"),
		sm.Where(afterOrEqualToday),
	).MustBuild(ctx)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		rlog.Error("Error querying reminders from db", "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	reminders := make([]Reminder, 0)
	for rows.Next() {
		var row Reminder
		if err := rows.Scan(&row.ID, &row.NextAlertDate, &row.AlertBefore); err != nil {
			rlog.Error("Error scanning reminder rows to array", "err", err.Error())
			return nil, err
		}
		reminders = append(reminders, row)
	}

	return reminders, nil
}

func GetRepeatableReminders(ctx context.Context, db DBTX) ([]Reminder, error) {
	today := time.Now().UTC().Format("2006-01-02")
	beforeToday := psql.Quote("next_alert_date").LT(psql.Arg(today))
	repeatYearly := psql.Quote("repeat").EQ(psql.Arg(RepeatModeYearly))
	repeatMonthly := psql.Quote("repeat").EQ(psql.Arg(RepeatModeMonthly))
	isRepeatable := psql.Or(repeatYearly, repeatMonthly)

	query, args := psql.Select(
		sm.Columns("id", "reminder_date", "repeat", "alert_before"),
		sm.From("reminder"),
		sm.Where(psql.And(beforeToday, isRepeatable)),
	).MustBuild(ctx)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		rlog.Error("Error querying reminders from db", "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	reminders := make([]Reminder, 0)
	for rows.Next() {
		var row Reminder
		if err := rows.Scan(&row.ID, &row.ReminderDate, &row.Repeat, &row.AlertBefore); err != nil {
			rlog.Error("Error scanning reminder rows to array", "err", err.Error())
			return nil, err
		}
		reminders = append(reminders, row)
	}

	return reminders, nil
}

func UpdateReminderNextAlertDate(ctx context.Context, db DBTX, reminderId string, nextAlertDate time.Time) error {
	query, args := psql.Update(
		um.Table("reminder"),
		um.SetCol("next_alert_date").To(psql.Arg(nextAlertDate)),
		um.Where(psql.Quote("id").EQ(psql.Arg(reminderId))),
	).MustBuild(ctx)

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		rlog.Error("Error updating reminder next alert date", "err", err.Error(), "reminderId", reminderId)
		return err
	}

	return nil
}
