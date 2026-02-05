package request

import "errors"

// Validation errors.
var (
	ErrUsernameRequired      = errors.New("username is required")
	ErrPasswordRequired      = errors.New("password is required")
	ErrScheduleRequired      = errors.New("schedule is required for daily inventories")
	ErrDateRequired          = errors.New("date is required")
	ErrItemIDRequired        = errors.New("item_id is required")
	ErrRealValueRequired     = errors.New("real_value is required")
	ErrInvalidSchedule       = errors.New("invalid schedule value")
	ErrInvalidDate           = errors.New("invalid date format")
	ErrInventoryTypeRequired = errors.New("inventory_type is required")
	ErrInvalidInventoryType  = errors.New("invalid inventory_type value")
)
