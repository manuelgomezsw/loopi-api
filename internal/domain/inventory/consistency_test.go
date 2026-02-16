package inventory

import (
	"testing"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// TestDiscrepancyConsistency verifies that for a fixed set of inventory details,
// the same "expected", "has discrepancy" and "difference" results are coherent
// and would yield the same counts used by GetDiscrepancies, CompleteInventory,
// GetDashboard and admin list (single source of truth).
func TestDiscrepancyConsistency(t *testing.T) {
	// One fixture: details that mix no-discrepancy and discrepancy cases.
	details := []*entity.InventoryDetail{
		{ItemID: 1, SuggestedValue: ptrU16(10), RealValue: ptrU16(10)},                                           // no diff
		{ItemID: 2, SuggestedValue: ptrU16(42), UnitsSold: ptrU16(3), RealValue: ptrU16(39)},                      // expected 39, no diff (regression case)
		{ItemID: 3, SuggestedValue: ptrU16(20), StockReceived: ptrU16(5), UnitsSold: ptrU16(2), RealValue: ptrU16(22)}, // expected 23, real 22 -> diff -1
		{ItemID: 4, SuggestedValue: ptrU16(8), RealValue: ptrU16(12)},                                           // expected 8, real 12 -> diff +4
		{ItemID: 5, SuggestedValue: ptrU16(0), RealValue: ptrU16(0)},                                             // no diff
	}

	// Compute as GetDiscrepancies / CompleteInventory / dashboard would: ExpectedAtEnd + HasDiscrepancy + Difference.
	var discrepancyCount int
	for _, d := range details {
		expected := ExpectedAtEnd(d)
		hasDisc := HasDiscrepancyFromExpectedEnd(d, expected)
		diff := DifferenceFromExpected(d, expected)

		// Coherence: hasDiscrepancy <=> diff != 0
		if hasDisc && diff == 0 {
			t.Errorf("item %d: HasDiscrepancy=true but Difference=0", d.ItemID)
		}
		if !hasDisc && diff != 0 {
			t.Errorf("item %d: HasDiscrepancy=false but Difference=%d", d.ItemID, diff)
		}

		// Difference must equal real - expected when real is set
		if d.RealValue != nil {
			wantDiff := int16(*d.RealValue) - int16(expected)
			if diff != wantDiff {
				t.Errorf("item %d: DifferenceFromExpected = %d, want %d (real - expected)", d.ItemID, diff, wantDiff)
			}
		}

		if hasDisc {
			discrepancyCount++
		}
	}

	// Same count whether we "list discrepancies" (len of filtered) or "count for complete" (discrepancyCount).
	// Simulate GetDiscrepancies filter:
	var listLen int
	for _, d := range details {
		expected := ExpectedAtEnd(d)
		if HasDiscrepancyFromExpectedEnd(d, expected) {
			listLen++
		}
	}
	if listLen != discrepancyCount {
		t.Errorf("discrepancy list length = %d, count = %d (must match)", listLen, discrepancyCount)
	}

	// Expected: items 3 and 4 have discrepancy (2 items).
	if discrepancyCount != 2 {
		t.Errorf("expected 2 details with discrepancy, got %d", discrepancyCount)
	}
}

// TestExpectedAtEndVsExpectedForAdminConsistency ensures that when admin view
// uses ExpectedForAdmin (with shrinkage), the same coherence holds: hasDiscrepancy
// and difference are consistent.
func TestExpectedAtEndVsExpectedForAdminConsistency(t *testing.T) {
	// Detail with shrinkage: suggested 10, shrinkage 1, received 0, sold 0 -> ExpectedForAdmin = 9.
	d := &entity.InventoryDetail{
		SuggestedValue: ptrU16(10),
		Shrinkage:      ptrU16(1),
		RealValue:      ptrU16(9),
	}
	expectedAtEnd := ExpectedAtEnd(d)       // 10 (no shrinkage in employee flow)
	expectedAdmin := ExpectedForAdmin(d)   // 9

	if expectedAtEnd != 10 || expectedAdmin != 9 {
		t.Errorf("ExpectedAtEnd = %d, ExpectedForAdmin = %d; want 10, 9", expectedAtEnd, expectedAdmin)
	}

	// With real=9: no discrepancy for admin (9==9), but discrepancy for employee flow (9!=10).
	if HasDiscrepancyFromExpectedEnd(d, expectedAtEnd) != true {
		t.Error("employee flow: real 9 != expected 10, should have discrepancy")
	}
	if HasDiscrepancyFromExpectedEnd(d, expectedAdmin) != false {
		t.Error("admin flow: real 9 == expected 9, should not have discrepancy")
	}
	if diff := DifferenceFromExpected(d, expectedAdmin); diff != 0 {
		t.Errorf("admin: difference should be 0, got %d", diff)
	}
}
