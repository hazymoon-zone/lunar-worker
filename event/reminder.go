package event

import (
	"context"
	"time"

	"github.com/6tail/lunar-go/calendar"
	"github.com/hazymoon22/lunar-worker/db"
	"github.com/hazymoon22/lunar-worker/lunar"
)

func checkAlertDateEligible(alertDate time.Time, alertBefore *int32) bool {
	now := time.Now().UTC()
	alertFrom := alertDate
	if alertBefore != nil {
		alertFrom = alertFrom.AddDate(0, 0, -int(*alertBefore))
	}

	return now.Equal(alertFrom) || now.After(alertFrom)
}

func GetNextAlertDate(repeat db.RepeatMode, reminderDate time.Time) *time.Time {
	return getNextAlertDateAt(repeat, reminderDate, time.Now().UTC())
}

func getNextAlertDateAt(repeat db.RepeatMode, reminderDate time.Time, atTime time.Time) *time.Time {
	lunarDate := calendar.NewLunarFromYmd(reminderDate.Year(), int(reminderDate.Month()), reminderDate.Day())
	if lunarDate == nil {
		return nil
	}

	solarDate := lunarDate.GetSolar()
	if solarDate == nil {
		return nil
	}

	lunarTimestamp := lunar.SolarToDate(*solarDate).UnixMilli()
	atTimestamp := atTime.UnixMilli()

	if lunarTimestamp >= atTimestamp {
		next := lunar.LunarToDate(*lunarDate)
		return &next
	}

	if repeat == db.RepeatModeYearly {
		return getNextAlertDateYearly(*lunarDate, atTime)
	}

	if repeat == db.RepeatModeMonthly {
		return getNextAlertDateMonthly(*lunarDate, atTime)
	}

	return nil
}

func getNextAlertDateYearly(reminderDate calendar.Lunar, atTime time.Time) *time.Time {
	lunarCurrentYear := lunar.GetLunarCurrentYear(reminderDate, atTime)
	if lunarCurrentYear == nil {
		return nil
	}

	solarCurrentYear := lunarCurrentYear.GetSolar()
	if solarCurrentYear == nil {
		return nil
	}
	dateCurrentYear := lunar.SolarToDate(*solarCurrentYear)

	lunarTime := dateCurrentYear.UnixMilli()
	nowTime := atTime.UnixMilli()

	if lunarTime >= nowTime {
		return &dateCurrentYear
	}

	lunarNextYear := lunar.GetLunarNextYear(reminderDate, atTime)
	if lunarNextYear == nil {
		return nil
	}

	lunarNextYearDate := lunar.LunarToDate(*lunarNextYear)

	return &lunarNextYearDate
}

func getNextAlertDateMonthly(reminderDate calendar.Lunar, atTime time.Time) *time.Time {
	lunarCurrentMonth := lunar.GetLunarCurrentMonth(reminderDate, atTime)
	if lunarCurrentMonth == nil {
		return nil
	}

	lunarCurrentMonthDate := lunar.LunarToDate(*lunarCurrentMonth)

	return &lunarCurrentMonthDate
}

func GetEligibleReminders(ctx context.Context, dbConn db.DBTX) ([]db.Reminder, error) {
	reminders, err := db.GetRemindersFromToday(ctx, dbConn)
	if err != nil {
		return nil, err
	}

	eligibleReminders := make([]db.Reminder, 0)
	for _, reminder := range reminders {
		if !checkAlertDateEligible(reminder.NextAlertDate, reminder.AlertBefore) {
			continue
		}

		eligibleReminders = append(eligibleReminders, reminder)
	}

	return eligibleReminders, nil
}
