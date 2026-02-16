package inventory

import (
	"testing"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

func TestExpectedAtEnd(t *testing.T) {
	tests := []struct {
		name           string
		d              *entity.InventoryDetail
		wantExpected   uint16
	}{
		{
			name: "all nil",
			d:    &entity.InventoryDetail{},
			wantExpected: 0,
		},
		{
			name: "suggested only",
			d:    &entity.InventoryDetail{SuggestedValue: ptrU16(10)},
			wantExpected: 10,
		},
		{
			name: "suggested minus units_sold",
			d: &entity.InventoryDetail{
				SuggestedValue: ptrU16(10),
				UnitsSold:      ptrU16(3),
			},
			wantExpected: 7,
		},
		{
			name: "suggested plus stock_received minus units_sold",
			d: &entity.InventoryDetail{
				SuggestedValue: ptrU16(10),
				StockReceived:  ptrU16(5),
				UnitsSold:      ptrU16(3),
			},
			wantExpected: 12,
		},
		{
			name: "negative clamped to zero",
			d: &entity.InventoryDetail{
				SuggestedValue: ptrU16(2),
				UnitsSold:      ptrU16(5),
			},
			wantExpected: 0,
		},
		{
			name: "regression: suggested 42 sold 3 -> 39",
			d: &entity.InventoryDetail{
				SuggestedValue: ptrU16(42),
				UnitsSold:      ptrU16(3),
			},
			wantExpected: 39,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpectedAtEnd(tt.d)
			if got != tt.wantExpected {
				t.Errorf("ExpectedAtEnd() = %v, want %v", got, tt.wantExpected)
			}
		})
	}
}

func TestExpectedForAdmin(t *testing.T) {
	tests := []struct {
		name         string
		d            *entity.InventoryDetail
		wantExpected uint16
	}{
		{
			name:         "all nil",
			d:            &entity.InventoryDetail{},
			wantExpected: 0,
		},
		{
			name: "suggested minus shrinkage",
			d: &entity.InventoryDetail{
				SuggestedValue: ptrU16(10),
				Shrinkage:      ptrU16(1),
			},
			wantExpected: 9,
		},
		{
			name: "suggested minus shrinkage plus received minus sold",
			d: &entity.InventoryDetail{
				SuggestedValue: ptrU16(10),
				Shrinkage:      ptrU16(1),
				StockReceived:  ptrU16(2),
				UnitsSold:      ptrU16(1),
			},
			wantExpected: 10,
		},
		{
			name: "negative clamped to zero",
			d: &entity.InventoryDetail{
				SuggestedValue: ptrU16(2),
				Shrinkage:      ptrU16(5),
			},
			wantExpected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpectedForAdmin(tt.d)
			if got != tt.wantExpected {
				t.Errorf("ExpectedForAdmin() = %v, want %v", got, tt.wantExpected)
			}
		})
	}
}

func ptrU16(v uint16) *uint16 {
	return &v
}
