package usecase

import (
	"errors"
	"loopi-api/internal/calendar"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/dto"
	"loopi-api/internal/usecase/utils"
)

type shiftProjectionUseCase struct {
	shiftRepo      repository.ShiftRepository
	workConfigRepo repository.WorkConfigRepository
}

type ShiftProjectionUseCase interface {
	PreviewHours(req dto.ShiftProjectionRequest) (domain.ExtraHourSummary, error)
}

func NewShiftProjectionUseCase(
	shiftRepo repository.ShiftRepository,
	workConfigRepo repository.WorkConfigRepository,
) ShiftProjectionUseCase {
	return &shiftProjectionUseCase{
		shiftRepo:      shiftRepo,
		workConfigRepo: workConfigRepo,
	}
}

func (u *shiftProjectionUseCase) PreviewHours(req dto.ShiftProjectionRequest) (domain.ExtraHourSummary, error) {
	shift, err := u.shiftRepo.GetByID(req.ShiftID)
	if err != nil || shift == nil {
		return domain.ExtraHourSummary{}, errors.New("invalid shift")
	}

	if req.Year <= 0 || req.Month < 1 || req.Month > 12 {
		return domain.ExtraHourSummary{}, errors.New("invalid period")
	}

	holidayMap := utils.HolidaysToMap(
		calendar.GetColombianHolidaysByMonthCached(req.Year, req.Month),
	)

	workConfig := u.workConfigRepo.GetActiveConfig()
	calendarDays := utils.BuildCalendarDays(req.Year, req.Month, holidayMap)

	projected := utils.ApplyShiftToCalendar(calendarDays, *shift, workConfig)
	summary := utils.SummarizeProjection(projected)
	summary.Period = domain.Period{
		Year:  req.Year,
		Month: req.Month,
	}
	return summary, nil
}
