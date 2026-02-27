package event

import (
	"testing"
	"time"

	"github.com/6tail/lunar-go/calendar"
	"github.com/hazymoon22/lunar-worker/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAlertDateEligibleWithoutAlertBefore(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	assert.True(t, checkAlertDateEligible(past, nil))

	future := time.Now().Add(1 * time.Hour)
	assert.False(t, checkAlertDateEligible(future, nil))
}

func TestCheckAlertDateEligibleWithAlertBefore(t *testing.T) {
	alertBefore := int32(2)

	// alertFrom is in the future (72h - 2 days = +24h from now)
	alertDateFuture := time.Now().Add(72 * time.Hour)
	assert.False(t, checkAlertDateEligible(alertDateFuture, &alertBefore))

	// alertFrom is in the past
	alertDateEligible := time.Now().Add(1 * time.Hour)
	assert.True(t, checkAlertDateEligible(alertDateEligible, &alertBefore))
}

func TestSolarToDate(t *testing.T) {
	solar := calendar.NewSolar(2026, 2, 27, 15, 4, 5)
	require.NotNil(t, solar)

	got := solarToDate(*solar)

	assert.Equal(t, 2026, got.Year())
	assert.Equal(t, time.February, got.Month())
	assert.Equal(t, 27, got.Day())
	assert.Equal(t, 0, got.Hour())
	assert.Equal(t, 0, got.Minute())
	assert.Equal(t, 0, got.Second())
	assert.Equal(t, 0, got.Nanosecond())
	assert.Equal(t, time.UTC, got.Location())
}

func TestGetNextAlertDateUnsupportedRepeatReturnsNil(t *testing.T) {
	reminderDate := time.Now().AddDate(-1, 0, 0)
	got := GetNextAlertDate(db.RepeatMode("none"), reminderDate)
	assert.Nil(t, got)
}

func TestGetNextAlertDateYearlyReturnsDate(t *testing.T) {
	reminderDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := GetNextAlertDate(db.RepeatModeYearly, reminderDate)
	assert.NotNil(t, got)
}

func TestGetNextAlertDateMonthlyReturnsDate(t *testing.T) {
	reminderDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := GetNextAlertDate(db.RepeatModeMonthly, reminderDate)
	assert.NotNil(t, got)
}
