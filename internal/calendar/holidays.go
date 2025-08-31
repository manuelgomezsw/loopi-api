package calendar

import "time"

// GetColombianHolidays Obtiene los días festivos Colombianos en un año específico
func GetColombianHolidays(year int) []time.Time {
	holidayMap := getColombianHolidays(year)

	var holidays []time.Time
	for date := range holidayMap {
		holidays = append(holidays, date)
	}
	return holidays
}

func getColombianHolidays(year int) map[time.Time]string {
	holidayMap := make(map[time.Time]string)
	easter := calculateEaster(year)

	// Fixed holidays
	fixed := map[time.Time]string{
		time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC):   "New Year's Day",
		time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC):   "Labor Day",
		time.Date(year, 7, 20, 0, 0, 0, 0, time.UTC):  "Independence Day",
		time.Date(year, 8, 7, 0, 0, 0, 0, time.UTC):   "Battle of Boyacá",
		time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC):  "Immaculate Conception",
		time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC): "Christmas",
	}
	for date, name := range fixed {
		holidayMap[date] = name
	}

	// Movable holidays (law 51/1983)
	movable := map[string]time.Time{
		"Epiphany":               time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC),
		"Saint Joseph's Day":     time.Date(year, 3, 19, 0, 0, 0, 0, time.UTC),
		"Saint Peter and Paul":   time.Date(year, 6, 29, 0, 0, 0, 0, time.UTC),
		"Assumption of Mary":     time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC),
		"Columbus Day":           time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC),
		"All Saints' Day":        time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC),
		"Cartagena Independence": time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC),
	}
	for name, date := range movable {
		observed := moveToMonday(date)
		holidayMap[observed] = name
	}

	// Easter-related holidays
	religious := map[string]time.Time{
		"Holy Thursday":      easter.AddDate(0, 0, -3),
		"Good Friday":        easter.AddDate(0, 0, -2),
		"Ascension of Jesus": moveToMonday(easter.AddDate(0, 0, 39)),
		"Corpus Christi":     moveToMonday(easter.AddDate(0, 0, 60)),
		"Sacred Heart":       moveToMonday(easter.AddDate(0, 0, 68)),
	}
	for name, date := range religious {
		holidayMap[date] = name
	}

	return holidayMap
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
func moveToMonday(d time.Time) time.Time {
	day := d.Weekday()
	if day == time.Sunday || (day >= time.Tuesday && day <= time.Saturday) {
		daysToAdd := (8 - int(day)) % 7
		return d.AddDate(0, 0, daysToAdd)
	}
	return d
}
