package inventory

import (
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// HasDiscrepancyFromExpectedEnd returns true when real_value != expected
// (e.g. when expected is ExpectedAtEnd(d) or ExpectedForAdmin(d)).
// If real_value is nil, returns false.
func HasDiscrepancyFromExpectedEnd(d *entity.InventoryDetail, expected uint16) bool {
	if d.RealValue == nil {
		return false
	}
	return *d.RealValue != expected
}

// DifferenceFromExpected returns real_value - expected (for display).
// If real_value is nil, returns 0.
func DifferenceFromExpected(d *entity.InventoryDetail, expected uint16) int16 {
	if d.RealValue == nil {
		return 0
	}
	return int16(*d.RealValue) - int16(expected)
}
