package lunar

import (
	"testing"
	"time"

	"github.com/6tail/lunar-go/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLunarCurrentYear(t *testing.T) {
	base := calendar.NewLunarFromYmd(2024, 1, 1)
	require.NotNil(t, base)
	atTime := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	got := GetLunarCurrentYear(*base, atTime)
	require.NotNil(t, got)

	assert.Equal(t, 2030, got.GetYear())
	assert.Equal(t, 1, got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}

func TestGetLunarNextYear(t *testing.T) {
	base := calendar.NewLunarFromYmd(2024, 1, 1)
	require.NotNil(t, base)
	atTime := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	got := GetLunarNextYear(*base, atTime)
	require.NotNil(t, got)

	assert.Equal(t, 2031, got.GetYear())
	assert.Equal(t, 1, got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}

func TestGetLunarCurrentMonth(t *testing.T) {
	base := calendar.NewLunarFromYmd(2024, 1, 1)
	require.NotNil(t, base)
	atTime := time.Date(2030, 2, 20, 0, 0, 0, 0, time.UTC)

	got := GetLunarCurrentMonth(*base, atTime)
	require.NotNil(t, got)

	assert.Equal(t, 2030, got.GetYear())
	assert.Equal(t, 2, got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}

func TestGetLunarNextMonth(t *testing.T) {
	base := calendar.NewLunarFromYmd(2024, 1, 1)
	require.NotNil(t, base)
	atTime := time.Date(2030, 2, 20, 0, 0, 0, 0, time.UTC)

	got := GetLunarNextMonth(*base, atTime)

	require.NotNil(t, got)
	assert.Equal(t, 2030, got.GetYear())
	assert.Equal(t, 3, got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}
