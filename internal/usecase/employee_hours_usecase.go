package usecase

import (
	"errors"
	"loopi-api/internal/calendar"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/utils"
)

type EmployeeHoursUseCase interface {
	GetMonthlySummary(employeeID, year, month int) (domain.EmployeeHourSummary, error)
}

type employeeHoursUseCase struct {
	assignedRepo repository.AssignedShiftRepository
	absenceRepo  repository.AbsenceRepository
	noveltyRepo  repository.NoveltyRepository
	userRepo     repository.UserRepository
}

func NewEmployeeHoursUseCase(
	assignedRepo repository.AssignedShiftRepository,
	absenceRepo repository.AbsenceRepository,
	noveltyRepo repository.NoveltyRepository,
	userRepo repository.UserRepository,
) EmployeeHoursUseCase {
	return &employeeHoursUseCase{
		assignedRepo: assignedRepo,
		absenceRepo:  absenceRepo,
		noveltyRepo:  noveltyRepo,
		userRepo:     userRepo,
	}
}

func (u *employeeHoursUseCase) GetMonthlySummary(employeeID, year, month int) (domain.EmployeeHourSummary, error) {
	if employeeID == 0 || year < 2000 || month < 1 || month > 12 {
		return domain.EmployeeHourSummary{}, errors.New("invalid parameters")
	}

	calendarDays := utils.BuildCalendarDays(year, month, utils.HolidaysToMap(calendar.GetColombianHolidaysByMonthCached(year, month)))
	shifts, _ := u.assignedRepo.GetByEmployeeAndMonth(employeeID, year, month)
	absences, _ := u.absenceRepo.GetByEmployeeAndMonth(employeeID, year, month)
	novelties, _ := u.noveltyRepo.GetByEmployeeAndMonth(employeeID, year, month)

	absenceMap := make(map[string]float64)
	noveltyMap := make(map[string]float64)
	for _, a := range absences {
		key := a.Date.Format("2006-01-02")
		absenceMap[key] += a.Hours
	}

	for _, n := range novelties {
		key := n.Date.Format("2006-01-02")
		if n.Type == "positive" {
			noveltyMap[key] += n.Hours
		} else {
			noveltyMap[key] -= n.Hours
		}
	}

	assignedMap := make(map[string]domain.AssignedShift)
	for _, s := range shifts {
		assignedMap[s.Date] = s
	}

	fullNameEmployee, _ := u.userRepo.GetNameByID(employeeID)

	summary := domain.EmployeeHourSummary{
		Employee: domain.EmployeeInfo{
			ID:       employeeID,
			FullName: fullNameEmployee,
		},
		Period: domain.Period{Year: year, Month: month},
	}

	for _, day := range calendarDays {
		d := day.Date.Format("2006-01-02")
		shift, ok := assignedMap[d]
		if !ok {
			continue // no shift
		}

		start := utils.ParseHour(shift.StartTime)
		end := utils.ParseHour(shift.EndTime)
		worked := utils.DurationInHours(start, end) - float64(shift.LunchMinutes)/60.0

		// aplicar novedades
		worked += noveltyMap[d]
		absence := absenceMap[d]

		extra := worked - 7.33
		if extra < 0 {
			extra = 0
		}

		diurnal, nocturnal := utils.SplitByFranja(start, end, utils.ParseHour("06:00"), utils.ParseHour("21:00"))
		scale := extra / (diurnal + nocturnal)
		diurnalExtra := utils.RoundTo2(scale * diurnal)
		nocturnalExtra := utils.RoundTo2(scale * nocturnal)

		block := &summary.Ordinary
		switch day.DayType {
		case utils.Sunday:
			block = &summary.Sunday
		case utils.Holiday:
			block = &summary.Holiday
		}

		block.Absence += utils.RoundTo2(absence)
		block.Novelty += utils.RoundTo2(noveltyMap[d])
		block.DiurnalExtra += diurnalExtra
		block.NocturnalExtra += nocturnalExtra
	}

	return summary, nil
}
