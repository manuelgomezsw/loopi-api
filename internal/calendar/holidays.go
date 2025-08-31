package calendar

import "time"

// GetColombianHolidays Obtiene los días festivos Colombianos en un año específico
func GetColombianHolidays(year int) []time.Time {
	holidays := []time.Time{}

	fixedDates := []time.Time{
		time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year, 7, 20, 0, 0, 0, 0, time.UTC),
		time.Date(year, 8, 7, 0, 0, 0, 0, time.UTC),
		time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC),
		time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC),
	}
	holidays = append(holidays, fixedDates...)

	moveToMonday := []time.Time{
		time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC),
		time.Date(year, 3, 19, 0, 0, 0, 0, time.UTC),
		time.Date(year, 6, 29, 0, 0, 0, 0, time.UTC),
		time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC),
		time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC),
		time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC),
	}
	for _, d := range moveToMonday {
		holidays = append(holidays, nextMonday(d))
	}

	easter := calculateEaster(year)
	holidays = append(holidays,
		easter.AddDate(0, 0, -3),             // Holy Thursday
		easter.AddDate(0, 0, -2),             // Good Friday
		nextMonday(easter.AddDate(0, 0, 43)), // Ascension
		nextMonday(easter.AddDate(0, 0, 64)), // Corpus Christi
		nextMonday(easter.AddDate(0, 0, 71)), // Sacred Heart
	)

	return holidays
}

// calculateEaster Calcula los días festivos de pascua
func calculateEaster(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// nextMonday Calcula los festivos que por ley se trasladan para el siguiente lunes
func nextMonday(d time.Time) time.Time {
	return d.AddDate(0, 0, (7-int(d.Weekday()))%7)
}
