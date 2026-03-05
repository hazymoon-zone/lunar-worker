package lunar

import (
	"time"

	"github.com/6tail/lunar-go/calendar"
)

func GetLunarCurrentYear(lunar calendar.Lunar, atTime time.Time) *calendar.Lunar {
	year := atTime.UTC().Year()
	month := lunar.GetMonth()
	day := lunar.GetDay()

	return calendar.NewLunarFromYmd(year, month, day)
}

func GetLunarNextYear(lunar calendar.Lunar, atTime time.Time) *calendar.Lunar {
	year := atTime.UTC().Year()
	targetYear := year + 1
	month := lunar.GetMonth()
	day := lunar.GetDay()

	return calendar.NewLunarFromYmd(targetYear, month, day)
}

func GetLunarCurrentMonth(lunar calendar.Lunar, atTime time.Time) *calendar.Lunar {
	now := atTime.UTC()
	year := now.Year()
	month := now.Month()
	day := lunar.GetDay()

	return calendar.NewLunarFromYmd(year, int(month), day)
}

func GetLunarNextMonth(lunar calendar.Lunar, atTime time.Time) *calendar.Lunar {
	now := atTime.UTC()
	year := now.Year()
	month := int(now.Month())
	targetMonth := month + 1
	day := lunar.GetDay()

	return calendar.NewLunarFromYmd(year, targetMonth, day)
}
