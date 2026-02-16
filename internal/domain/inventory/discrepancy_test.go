package inventory

import (
	"testing"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

func TestHasDiscrepancyFromExpectedEnd(t *testing.T) {
	tests := []struct {
		name     string
		d        *entity.InventoryDetail
		expected uint16
		want     bool
	}{
		{
			name:     "real nil -> no discrepancy",
			d:        &entity.InventoryDetail{},
			expected: 10,
			want:     false,
		},
		{
			name:     "real equals expected -> no discrepancy",
			d:        &entity.InventoryDetail{RealValue: ptrU16(39)},
			expected: 39,
			want:     false,
		},
		{
			name:     "real differs from expected -> discrepancy",
			d:        &entity.InventoryDetail{RealValue: ptrU16(42)},
			expected: 39,
			want:     true,
		},
		{
			name:     "regression: suggested 42 real 39 sold 3 -> expected 39, no discrepancy",
			d:        &entity.InventoryDetail{RealValue: ptrU16(39), SuggestedValue: ptrU16(42), UnitsSold: ptrU16(3)},
			expected: 39,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasDiscrepancyFromExpectedEnd(tt.d, tt.expected)
			if got != tt.want {
				t.Errorf("HasDiscrepancyFromExpectedEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDifferenceFromExpected(t *testing.T) {
	tests := []struct {
		name     string
		d        *entity.InventoryDetail
		expected uint16
		want     int16
	}{
		{
			name:     "real nil -> 0",
			d:        &entity.InventoryDetail{},
			expected: 10,
			want:     0,
		},
		{
			name:     "real 39 expected 39 -> 0",
			d:        &entity.InventoryDetail{RealValue: ptrU16(39)},
			expected: 39,
			want:     0,
		},
		{
			name:     "real 42 expected 39 -> +3",
			d:        &entity.InventoryDetail{RealValue: ptrU16(42)},
			expected: 39,
			want:     3,
		},
		{
			name:     "real 36 expected 39 -> -3",
			d:        &entity.InventoryDetail{RealValue: ptrU16(36)},
			expected: 39,
			want:     -3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DifferenceFromExpected(tt.d, tt.expected)
			if got != tt.want {
				t.Errorf("DifferenceFromExpected() = %v, want %v", got, tt.want)
			}
		})
	}
}
