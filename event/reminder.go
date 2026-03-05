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

func solarToDate(solar calendar.Solar) time.Time {
	return time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 0, 0, 0, 0, time.UTC)
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

	lunarTimestamp := solarToDate(*solarDate).UnixMilli()
	atTimestamp := atTime.UnixMilli()

	if lunarTimestamp >= atTimestamp {
		next := solarToDate(*lunarDate.GetSolar())
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
	dateCurrentYear := solarToDate(*solarCurrentYear)

	lunarTime := dateCurrentYear.UnixMilli()
	nowTime := atTime.UnixMilli()

	if lunarTime >= nowTime {
		return &dateCurrentYear
	}

	lunarNextYear := lunar.GetLunarNextYear(reminderDate, atTime)
	if lunarNextYear == nil {
		return nil
	}
	solarNextYear := lunarNextYear.GetSolar()
	if solarNextYear == nil {
		return nil
	}
	dateNextYear := solarToDate(*solarNextYear)

	return &dateNextYear
}

func getNextAlertDateMonthly(reminderDate calendar.Lunar, atTime time.Time) *time.Time {
	lunarCurrentMonth := lunar.GetLunarCurrentMonth(reminderDate, atTime)
	if lunarCurrentMonth == nil {
		return nil
	}

	solarCurrentMonth := lunarCurrentMonth.GetSolar()
	if solarCurrentMonth == nil {
		return nil
	}
	dateCurrentMonth := solarToDate(*solarCurrentMonth)

	lunarTime := dateCurrentMonth.UnixMilli()
	nowTime := atTime.UnixMilli()

	if lunarTime >= nowTime {
		return &dateCurrentMonth
	}

	lunarNextMonth := lunar.GetLunarNextMonth(reminderDate, atTime)
	if lunarNextMonth == nil {
		return nil
	}
	solarNextMonth := lunarNextMonth.GetSolar()
	if solarNextMonth == nil {
		return nil
	}
	dateNextMonth := solarToDate(*solarNextMonth)

	return &dateNextMonth
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
