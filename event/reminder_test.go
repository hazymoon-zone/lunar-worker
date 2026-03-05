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
	reminderDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	atTime := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	got := getNextAlertDateAt(db.RepeatMode("none"), reminderDate, atTime)
	assert.Nil(t, got)
}

func TestGetNextAlertDateReturnsCurrentLunarDateWhenNotPassed(t *testing.T) {
	reminderDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	atTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	got := getNextAlertDateAt(db.RepeatModeYearly, reminderDate, atTime)
	require.NotNil(t, got)

	lunarDate := calendar.NewLunarFromYmd(reminderDate.Year(), int(reminderDate.Month()), reminderDate.Day())
	require.NotNil(t, lunarDate)
	expectedSolar := lunarDate.GetSolar()
	require.NotNil(t, expectedSolar)
	expected := solarToDate(*expectedSolar)

	assert.Equal(t, expected, *got)
}

func TestGetNextAlertDateYearlyFollowsYearlyBranch(t *testing.T) {
	reminderDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	atTime := time.Date(2030, 3, 1, 0, 0, 0, 0, time.UTC)

	got := getNextAlertDateAt(db.RepeatModeYearly, reminderDate, atTime)
	require.NotNil(t, got)

	expected := time.Date(2031, 1, 23, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, *got)
}

func TestGetNextAlertDateMonthlyFollowsMonthlyBranch(t *testing.T) {
	reminderDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	atTime := time.Date(2030, 2, 20, 0, 0, 0, 0, time.UTC)

	got := getNextAlertDateAt(db.RepeatModeMonthly, reminderDate, atTime)
	require.NotNil(t, got)

	expected := time.Date(2030, 3, 4, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, *got)
}
