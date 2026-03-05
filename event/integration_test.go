package event

import (
	"context"
	"os"
	"testing"
	"time"

	"encore.dev/types/uuid"
	"github.com/hazymoon22/lunar-worker/db"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if os.Getenv("RUN_EVENT_INTEGRATION") != "1" {
		t.Skip("set RUN_EVENT_INTEGRATION=1 to run event integration tests with Encore runtime")
	}

	pool, err := db.GetDatabasePool(context.Background())
	require.NoError(t, err)

	return pool
}

func withTx(t *testing.T, fn func(ctx context.Context, tx pgx.Tx)) {
	t.Helper()

	pool := testPool(t)
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	fn(ctx, tx)
}

func newUUIDString(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV4()
	require.NoError(t, err)
	return id.String()
}

func seedUser(t *testing.T, ctx context.Context, tx pgx.Tx, userID string) {
	t.Helper()
	_, err := tx.Exec(ctx, `INSERT INTO "user"(id, name, email) VALUES ($1, $2, $3)`, userID, "Test User", "test@example.com")
	require.NoError(t, err)
}

func seedReminder(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	id string,
	userID string,
	reminderDate time.Time,
	nextAlertDate time.Time,
	repeat *db.RepeatMode,
	alertBefore *int32,
) {
	t.Helper()
	_, err := tx.Exec(ctx, `
		INSERT INTO reminder(
			id, title, reminder_date, next_alert_date, user_id, repeat, alert_before, mail_subject, mail_body
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, "Reminder", reminderDate, nextAlertDate, userID, repeat, alertBefore, "Subject", "<p>Body</p>")
	require.NoError(t, err)
}

func seedAlert(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	id string,
	reminderID string,
	alertDate time.Time,
	acknowledged bool,
) {
	t.Helper()
	_, err := tx.Exec(ctx, `
		INSERT INTO alert(id, reminder_id, alert_date, acknowledged)
		VALUES ($1, $2, $3, $4)
	`, id, reminderID, alertDate, acknowledged)
	require.NoError(t, err)
}

func TestIntegrationGetEligibleReminders(t *testing.T) {
	withTx(t, func(ctx context.Context, tx pgx.Tx) {
		userID := "user-get-eligible-reminders"
		seedUser(t, ctx, tx, userID)

		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		tomorrow := today.AddDate(0, 0, 1)
		yesterday := today.AddDate(0, 0, -1)
		yearly := db.RepeatModeYearly

		eligibleTodayID := newUUIDString(t)
		eligibleWithLeadID := newUUIDString(t)
		ineligibleFutureID := newUUIDString(t)
		pastID := newUUIDString(t)
		alertBeforeTwoDays := int32(2)

		seedReminder(t, ctx, tx, eligibleTodayID, userID, today, today, &yearly, nil)
		seedReminder(t, ctx, tx, eligibleWithLeadID, userID, tomorrow, tomorrow, &yearly, &alertBeforeTwoDays)
		seedReminder(t, ctx, tx, ineligibleFutureID, userID, tomorrow, tomorrow, &yearly, nil)
		seedReminder(t, ctx, tx, pastID, userID, yesterday, yesterday, &yearly, nil)

		eligible, err := GetEligibleReminders(ctx, tx)
		require.NoError(t, err)

		ids := make(map[string]struct{}, len(eligible))
		for _, r := range eligible {
			ids[r.ID] = struct{}{}
		}

		_, hasEligibleToday := ids[eligibleTodayID]
		_, hasEligibleWithLead := ids[eligibleWithLeadID]
		_, hasIneligibleFuture := ids[ineligibleFutureID]
		_, hasPast := ids[pastID]

		assert.True(t, hasEligibleToday)
		assert.True(t, hasEligibleWithLead)
		assert.False(t, hasIneligibleFuture)
		assert.False(t, hasPast)
	})
}

func TestIntegrationCreateAlertsForEligibleReminders(t *testing.T) {
	withTx(t, func(ctx context.Context, tx pgx.Tx) {
		userID := "user-create-alerts"
		seedUser(t, ctx, tx, userID)

		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		yearly := db.RepeatModeYearly

		reminderExistingID := newUUIDString(t)
		reminderNewID := newUUIDString(t)
		seedReminder(t, ctx, tx, reminderExistingID, userID, today, today, &yearly, nil)
		seedReminder(t, ctx, tx, reminderNewID, userID, today, today, &yearly, nil)

		existingAlertID := newUUIDString(t)
		seedAlert(t, ctx, tx, existingAlertID, reminderExistingID, today, false)

		eligibleReminders := []db.Reminder{
			{ID: reminderExistingID},
			{ID: reminderNewID},
		}

		created, err := CreateAlertsForEligibleReminders(ctx, tx, eligibleReminders)
		require.NoError(t, err)
		require.Len(t, created, 1)
		assert.Equal(t, reminderNewID, created[0].ReminderID)

		var totalToday int
		err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM alert WHERE alert_date = $1`, today).Scan(&totalToday)
		require.NoError(t, err)
		assert.Equal(t, 2, totalToday)
	})
}
