package api

import (
	"context"
	"os"
	"testing"
	"time"

	"encore.dev/types/uuid"
	"github.com/hazymoon22/lunar-worker/acktoken"
	"github.com/hazymoon22/lunar-worker/db"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if os.Getenv("RUN_API_INTEGRATION") != "1" {
		t.Skip("set RUN_API_INTEGRATION=1 to run api integration tests with Encore runtime")
	}

	pool, err := db.GetDatabasePool(context.Background())
	require.NoError(t, err)

	return pool
}

func newTestUUID(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV4()
	require.NoError(t, err)
	return id.String()
}

func seedTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID string) {
	t.Helper()
	_, err := pool.Exec(ctx, `INSERT INTO "user"(id, name, email) VALUES ($1, $2, $3)`, userID, "API Test User", "api-test@example.com")
	require.NoError(t, err)
}

func seedTestReminder(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
	userID string,
	reminderDate time.Time,
	nextAlertDate time.Time,
	repeat string,
	alertBefore *int32,
) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO reminder(
			id, title, reminder_date, next_alert_date, user_id, repeat, alert_before, mail_subject, mail_body
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, "API Reminder", reminderDate, nextAlertDate, userID, repeat, alertBefore, "Subject", "<p>Body</p>")
	require.NoError(t, err)
}

func seedTestAlert(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
	reminderID string,
	alertDate time.Time,
	acknowledged bool,
) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO alert(id, reminder_id, alert_date, acknowledged)
		VALUES ($1, $2, $3, $4)
	`, id, reminderID, alertDate, acknowledged)
	require.NoError(t, err)
}

func TestIntegrationManageAlertsApi(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)

	userID := "api-manage-alerts-user"
	reminderID := newTestUUID(t)
	expiredAlertID := newTestUUID(t)
	todayAlertDate := time.Now().UTC()
	yesterday := todayAlertDate.AddDate(0, 0, -1)

	seedTestUser(t, ctx, pool, userID)
	seedTestReminder(t, ctx, pool, reminderID, userID, todayAlertDate, todayAlertDate, "", nil)
	seedTestAlert(t, ctx, pool, expiredAlertID, reminderID, yesterday, false)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM alert WHERE reminder_id = $1`, reminderID)
		_, _ = pool.Exec(ctx, `DELETE FROM reminder WHERE id = $1`, reminderID)
		_, _ = pool.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, userID)
	})

	err := ManageAlertsApi(ctx)
	require.NoError(t, err)

	var expiredCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM alert WHERE id = $1`, expiredAlertID).Scan(&expiredCount)
	require.NoError(t, err)
	assert.Equal(t, 0, expiredCount)

	var todayCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM alert
		WHERE reminder_id = $1
		  AND alert_date = $2::date
	`, reminderID, todayAlertDate).Scan(&todayCount)
	require.NoError(t, err)
	assert.Equal(t, 1, todayCount)
}

func TestIntegrationRenewRepeatableRemindersApi(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)

	userID := "api-renew-reminders-user"
	reminderID := newTestUUID(t)
	reminderDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	pastNextAlertDate := time.Now().UTC().AddDate(0, 0, -1)

	seedTestUser(t, ctx, pool, userID)
	seedTestReminder(t, ctx, pool, reminderID, userID, reminderDate, pastNextAlertDate, "yearly", nil)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM alert WHERE reminder_id = $1`, reminderID)
		_, _ = pool.Exec(ctx, `DELETE FROM reminder WHERE id = $1`, reminderID)
		_, _ = pool.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, userID)
	})

	err := RenewRepeatableRemindersApi(ctx)
	require.NoError(t, err)

	var updated time.Time
	err = pool.QueryRow(ctx, `SELECT next_alert_date FROM reminder WHERE id = $1`, reminderID).Scan(&updated)
	require.NoError(t, err)
	assert.NotEqual(t, pastNextAlertDate.Format("2006-01-02"), updated.UTC().Format("2006-01-02"))
}

func TestIntegrationSendAlertsApi_NoAlerts(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)

	// Ensure no sendable alerts exist to keep this test independent from Mailgun.
	_, err := pool.Exec(ctx, `DELETE FROM alert WHERE alert_date = $1::date`, time.Now().UTC())
	require.NoError(t, err)

	err = SendAlertsApi(ctx)
	require.NoError(t, err)
}

func TestIntegrationAcknowledgeAlertApi_NilParams(t *testing.T) {
	ctx := context.Background()
	res, err := AcknowledgeAlertApi(ctx, nil)
	require.Error(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "", res.Message)
}

func TestIntegrationAcknowledgeAlertApi_InvalidToken(t *testing.T) {
	ctx := context.Background()
	res, err := AcknowledgeAlertApi(ctx, &AcknowledgeAlertQueryParams{Token: "invalid-token"})
	require.Error(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "", res.Message)
}

func TestIntegrationAcknowledgeAlertApi_ValidToken(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)

	userID := "api-ack-valid-user"
	reminderID := newTestUUID(t)
	alertID := newTestUUID(t)
	today := time.Now().UTC()

	seedTestUser(t, ctx, pool, userID)
	seedTestReminder(t, ctx, pool, reminderID, userID, today, today, "", nil)
	seedTestAlert(t, ctx, pool, alertID, reminderID, today, false)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM alert WHERE id = $1`, alertID)
		_, _ = pool.Exec(ctx, `DELETE FROM reminder WHERE id = $1`, reminderID)
		_, _ = pool.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, userID)
	})

	token, err := acktoken.Generate(alertID)
	require.NoError(t, err)

	res, err := AcknowledgeAlertApi(ctx, &AcknowledgeAlertQueryParams{Token: token})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "Reminder acknowledged", res.Message)

	var acknowledged bool
	err = pool.QueryRow(ctx, `SELECT acknowledged FROM alert WHERE id = $1`, alertID).Scan(&acknowledged)
	require.NoError(t, err)
	assert.True(t, acknowledged)
}
