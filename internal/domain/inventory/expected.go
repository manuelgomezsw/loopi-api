package inventory

import (
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// ExpectedAtEnd returns the expected count at end of period for employee/summary flow:
// suggested + stock_received - units_sold, clamped to >= 0.
// Shrinkage is not subtracted in this formula.
func ExpectedAtEnd(d *entity.InventoryDetail) uint16 {
	suggested := uint16(0)
	if d.SuggestedValue != nil {
		suggested = *d.SuggestedValue
	}
	unitsSold := uint16(0)
	if d.UnitsSold != nil {
		unitsSold = *d.UnitsSold
	}
	stockReceived := uint16(0)
	if d.StockReceived != nil {
		stockReceived = *d.StockReceived
	}
	expected := int32(suggested) + int32(stockReceived) - int32(unitsSold)
	if expected < 0 {
		expected = 0
	}
	return uint16(expected)
}

// ExpectedForAdmin returns the expected count for the admin detail view:
// suggested - shrinkage + stock_received - units_sold, clamped to >= 0.
// Used when the admin view must reflect shrinkage in the expected value.
func ExpectedForAdmin(d *entity.InventoryDetail) uint16 {
	suggested := uint16(0)
	if d.SuggestedValue != nil {
		suggested = *d.SuggestedValue
	}
	shrinkage := uint16(0)
	if d.Shrinkage != nil {
		shrinkage = *d.Shrinkage
	}
	stockReceived := uint16(0)
	if d.StockReceived != nil {
		stockReceived = *d.StockReceived
	}
	unitsSold := uint16(0)
	if d.UnitsSold != nil {
		unitsSold = *d.UnitsSold
	}
	expected := int32(suggested) - int32(shrinkage) + int32(stockReceived) - int32(unitsSold)
	if expected < 0 {
		expected = 0
	}
	return uint16(expected)
}
