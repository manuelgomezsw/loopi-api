package utils

import (
	"fmt"
	"loopi-api/internal/domain"
	"math"
	"time"
)

type projectedDay struct {
	Date           time.Time
	Type           DayType
	DiurnalExtra   float64
	NocturnalExtra float64
}

func ApplyShiftToCalendar(
	days []CalendarDay,
	shift domain.Shift,
	config domain.WorkConfig,
) []projectedDay {
	shiftStart := parseHour(shift.StartTime)
	shiftEnd := parseHour(shift.EndTime)
	diurnalStart := parseHour(config.DiurnalStart)
	diurnalEnd := parseHour(config.DiurnalEnd)
	dailyLimit := 7.33

	var result []projectedDay

	for _, day := range days {
		totalWorked := durationInHours(shiftStart, shiftEnd) - float64(shift.LunchMinutes)/60.0 // Se restan los minutos de almuerzo
		logTotalWorkedPerDay(day.Date, totalWorked)
		if totalWorked <= dailyLimit {
			continue // no extra
		}

		extra := totalWorked - dailyLimit
		diurnal, nocturnal := splitByFranja(shiftStart, shiftEnd, diurnalStart, diurnalEnd)

		scale := extra / (diurnal + nocturnal)
		diurnalExtra := roundTo2(scale * diurnal)
		nocturnalExtra := roundTo2(scale * nocturnal)

		result = append(result, projectedDay{
			Date:           day.Date,
			Type:           day.DayType,
			DiurnalExtra:   diurnalExtra,
			NocturnalExtra: nocturnalExtra,
		})
	}

	return result
}

func SummarizeProjection(days []projectedDay) domain.ExtraHourSummary {
	var summary domain.ExtraHourSummary

	for _, d := range days {
		switch d.Type {
		case Ordinary:
			summary.Ordinary.DiurnalExtra += d.DiurnalExtra
			summary.Ordinary.NocturnalExtra += d.NocturnalExtra
		case Sunday:
			summary.Sunday.DiurnalExtra += d.DiurnalExtra
			summary.Sunday.NocturnalExtra += d.NocturnalExtra
		case Holiday:
			summary.Holiday.DiurnalExtra += d.DiurnalExtra
			summary.Holiday.NocturnalExtra += d.NocturnalExtra
		}
	}

	summary.Ordinary.DiurnalExtra = roundTo2(summary.Ordinary.DiurnalExtra)
	summary.Ordinary.NocturnalExtra = roundTo2(summary.Ordinary.NocturnalExtra)
	summary.Sunday.DiurnalExtra = roundTo2(summary.Sunday.DiurnalExtra)
	summary.Sunday.NocturnalExtra = roundTo2(summary.Sunday.NocturnalExtra)
	summary.Holiday.DiurnalExtra = roundTo2(summary.Holiday.DiurnalExtra)
	summary.Holiday.NocturnalExtra = roundTo2(summary.Holiday.NocturnalExtra)

	return summary
}

func parseHour(value string) time.Time {
	layouts := []string{"15:04", "15:04:05"}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return time.Date(2000, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
		}
	}

	// Fallback a 00:00 si falla
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
}

func durationInHours(start, end time.Time) float64 {
	if end.Before(start) || end.Equal(start) {
		end = end.Add(24 * time.Hour)
	}
	dur := end.Sub(start).Hours()
	return dur
}

func splitByFranja(start, end, diurnalStart, diurnalEnd time.Time) (float64, float64) {
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}
	diurnal := 0.0
	nocturnal := 0.0

	for t := start; t.Before(end); t = t.Add(1 * time.Minute) {
		isDiurnal := t.After(diurnalStart) && t.Before(diurnalEnd)
		if isDiurnal {
			diurnal += 1.0 / 60
		} else {
			nocturnal += 1.0 / 60
		}
	}
	return diurnal, nocturnal
}

func roundTo2(val float64) float64 {
	return math.Round(val*100) / 100
}

func logTotalWorkedPerDay(day time.Time, worked float64) {
	fmt.Printf("[DEBUG] %s - Worked: %.2f\n", day.Format("2006-01-02"), worked)
}
