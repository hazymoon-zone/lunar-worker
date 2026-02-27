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

	got := GetLunarCurrentYear(*base)
	require.NotNil(t, got)

	now := time.Now()
	assert.Equal(t, now.Year(), got.GetYear())
	assert.Equal(t, 1, got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}

func TestGetLunarNextYear(t *testing.T) {
	base := calendar.NewLunarFromYmd(2024, 1, 1)
	require.NotNil(t, base)

	got := GetLunarNextYear(*base)
	require.NotNil(t, got)

	now := time.Now()
	assert.Equal(t, now.Year()+1, got.GetYear())
	assert.Equal(t, 1, got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}

func TestGetLunarCurrentMonth(t *testing.T) {
	base := calendar.NewLunarFromYmd(2024, 1, 1)
	require.NotNil(t, base)

	got := GetLunarCurrentMonth(*base)
	require.NotNil(t, got)

	now := time.Now()
	assert.Equal(t, now.Year(), got.GetYear())
	assert.Equal(t, int(now.Month()), got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}

func TestGetLunarNextMonth(t *testing.T) {
	base := calendar.NewLunarFromYmd(2024, 1, 1)
	require.NotNil(t, base)

	got := GetLunarNextMonth(*base)

	now := time.Now()
	if now.Month() == time.December {
		assert.Nil(t, got)
		return
	}

	require.NotNil(t, got)
	assert.Equal(t, now.Year(), got.GetYear())
	assert.Equal(t, int(now.Month())+1, got.GetMonth())
	assert.Equal(t, 1, got.GetDay())
}
