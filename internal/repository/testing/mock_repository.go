package testing

import (
	"errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/repository/mysql"
	"time"
)

// MockStoreRepository implements repository.StoreRepository for testing
type MockStoreRepository struct {
	stores     map[uint]domain.Store
	nextID     uint
	shouldFail bool
}

// NewMockStoreRepository creates a new mock store repository
func NewMockStoreRepository() *MockStoreRepository {
	return &MockStoreRepository{
		stores: make(map[uint]domain.Store),
		nextID: 1,
	}
}

// SetShouldFail configures the mock to simulate failures
func (m *MockStoreRepository) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// SeedData adds test data to the mock
func (m *MockStoreRepository) SeedData(stores []domain.Store) {
	for _, store := range stores {
		if store.ID == 0 {
			store.ID = m.nextID
			m.nextID++
		}
		m.stores[store.ID] = store
	}
}

// GetAll returns all stores
func (m *MockStoreRepository) GetAll() ([]domain.Store, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetAll failed")
	}

	stores := make([]domain.Store, 0, len(m.stores))
	for _, store := range m.stores {
		stores = append(stores, store)
	}
	return stores, nil
}

// GetByID returns a store by ID
func (m *MockStoreRepository) GetByID(id int) (domain.Store, error) {
	if m.shouldFail {
		return domain.Store{}, errors.New("mock error: GetByID failed")
	}

	store, exists := m.stores[uint(id)]
	if !exists {
		return domain.Store{}, mysql.ErrNotFound
	}
	return store, nil
}

// GetByFranchiseID returns stores by franchise ID
func (m *MockStoreRepository) GetByFranchiseID(franchiseID int) ([]domain.Store, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetByFranchiseID failed")
	}

	var stores []domain.Store
	for _, store := range m.stores {
		if store.FranchiseID == uint(franchiseID) {
			stores = append(stores, store)
		}
	}
	return stores, nil
}

// Create creates a new store
func (m *MockStoreRepository) Create(store *domain.Store) error {
	if m.shouldFail {
		return errors.New("mock error: Create failed")
	}

	if store.Name == "" {
		return mysql.ErrInvalidInput
	}

	store.ID = m.nextID
	m.nextID++
	m.stores[store.ID] = *store
	return nil
}

// Update updates an existing store
func (m *MockStoreRepository) Update(store *domain.Store) error {
	if m.shouldFail {
		return errors.New("mock error: Update failed")
	}

	if _, exists := m.stores[store.ID]; !exists {
		return mysql.ErrNotFound
	}

	m.stores[store.ID] = *store
	return nil
}

// Delete removes a store
func (m *MockStoreRepository) Delete(id int) error {
	if m.shouldFail {
		return errors.New("mock error: Delete failed")
	}

	if _, exists := m.stores[uint(id)]; !exists {
		return mysql.ErrNotFound
	}

	delete(m.stores, uint(id))
	return nil
}

// GetActiveStoresByFranchise returns only active stores for a franchise
func (m *MockStoreRepository) GetActiveStoresByFranchise(franchiseID int) ([]domain.Store, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetActiveStoresByFranchise failed")
	}

	var stores []domain.Store
	for _, store := range m.stores {
		if store.FranchiseID == uint(franchiseID) && store.IsActive {
			stores = append(stores, store)
		}
	}
	return stores, nil
}

// GetStoresWithEmployeeCount returns stores with their employee counts for a franchise
func (m *MockStoreRepository) GetStoresWithEmployeeCount(franchiseID int) ([]repository.StoreWithEmployeeCount, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetStoresWithEmployeeCount failed")
	}

	var storesWithCount []repository.StoreWithEmployeeCount
	for _, store := range m.stores {
		if store.FranchiseID == uint(franchiseID) && store.IsActive {
			// Mock employee count (in real implementation this would come from join)
			storeWithCount := repository.StoreWithEmployeeCount{
				Store:         store,
				EmployeeCount: int(store.ID * 3), // Mock calculation for testing
			}
			storesWithCount = append(storesWithCount, storeWithCount)
		}
	}
	return storesWithCount, nil
}

// GetStoreStatistics returns comprehensive statistics for a store
func (m *MockStoreRepository) GetStoreStatistics(storeID int) (*repository.StoreStatistics, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetStoreStatistics failed")
	}

	store, exists := m.stores[uint(storeID)]
	if !exists {
		return nil, mysql.ErrNotFound
	}

	// Mock statistics (in real implementation this would come from database queries)
	stats := &repository.StoreStatistics{
		Store:         store,
		EmployeeCount: int64(storeID * 5),  // Mock employee count
		ShiftCount:    int64(storeID * 10), // Mock shift count
	}

	return stats, nil
}

// RepositoryTestSuite provides common test utilities for repositories
type RepositoryTestSuite struct {
	StoreRepo   repository.StoreRepository
	ShiftRepo   repository.ShiftRepository
	AbsenceRepo repository.AbsenceRepository
	NoveltyRepo repository.NoveltyRepository
}

// NewRepositoryTestSuite creates a new test suite
func NewRepositoryTestSuite(useRealDB bool) *RepositoryTestSuite {
	var storeRepo repository.StoreRepository
	var shiftRepo repository.ShiftRepository
	var absenceRepo repository.AbsenceRepository
	var noveltyRepo repository.NoveltyRepository

	if useRealDB {
		// In real tests, you would inject a test database connection
		// storeRepo = mysql.NewStoreRepository(testDB)
		// shiftRepo = mysql.NewShiftRepository(testDB)
		// absenceRepo = mysql.NewAbsenceRepository(testDB)
		// noveltyRepo = mysql.NewNoveltyRepository(testDB)
		panic("Real DB testing not implemented in this example")
	} else {
		storeRepo = NewMockStoreRepository()
		shiftRepo = NewMockShiftRepository()
		absenceRepo = NewMockAbsenceRepository()
		noveltyRepo = NewMockNoveltyRepository()
	}

	return &RepositoryTestSuite{
		StoreRepo:   storeRepo,
		ShiftRepo:   shiftRepo,
		AbsenceRepo: absenceRepo,
		NoveltyRepo: noveltyRepo,
	}
}

// SeedTestData adds common test data
func (suite *RepositoryTestSuite) SeedTestData() {
	if mockStoreRepo, ok := suite.StoreRepo.(*MockStoreRepository); ok {
		testStores := []domain.Store{
			{Name: "Store 1", Code: "ST1", FranchiseID: 1, IsActive: true},
			{Name: "Store 2", Code: "ST2", FranchiseID: 1, IsActive: true},
			{Name: "Store 3", Code: "ST3", FranchiseID: 2, IsActive: false},
		}
		mockStoreRepo.SeedData(testStores)
	}

	if mockShiftRepo, ok := suite.ShiftRepo.(*MockShiftRepository); ok {
		testShifts := []domain.Shift{
			{Name: "Morning Shift", StoreID: 1, StartTime: "08:00", EndTime: "16:00", IsActive: true},
			{Name: "Evening Shift", StoreID: 1, StartTime: "16:00", EndTime: "00:00", IsActive: true},
			{Name: "Night Shift", StoreID: 1, StartTime: "00:00", EndTime: "08:00", IsActive: false},
			{Name: "Day Shift", StoreID: 2, StartTime: "09:00", EndTime: "17:00", IsActive: true},
		}
		mockShiftRepo.SeedShiftData(testShifts)
	}

	if mockAbsenceRepo, ok := suite.AbsenceRepo.(*MockAbsenceRepository); ok {
		testAbsences := []domain.Absence{
			{EmployeeID: 1, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Hours: 8.0, Reason: "Sick leave"},
			{EmployeeID: 1, Date: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), Hours: 4.0, Reason: "Medical appointment"},
			{EmployeeID: 2, Date: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Hours: 8.0, Reason: "Personal matters"},
		}
		mockAbsenceRepo.SeedAbsenceData(testAbsences)
	}

	if mockNoveltyRepo, ok := suite.NoveltyRepo.(*MockNoveltyRepository); ok {
		testNovelties := []domain.Novelty{
			{EmployeeID: 1, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Hours: 2.0, Type: "positive", Comment: "Extra hours"},
			{EmployeeID: 1, Date: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), Hours: 1.0, Type: "negative", Comment: "Late arrival"},
			{EmployeeID: 2, Date: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Hours: 3.0, Type: "positive", Comment: "Overtime"},
		}
		mockNoveltyRepo.SeedNoveltyData(testNovelties)
	}
}

// AssertStoreEqual compares two stores for equality
func (suite *RepositoryTestSuite) AssertStoreEqual(expected, actual domain.Store) bool {
	return expected.Name == actual.Name &&
		expected.FranchiseID == actual.FranchiseID &&
		expected.IsActive == actual.IsActive
}

// Example test cases (you would put these in actual *_test.go files)

// TestCreateStore demonstrates how to test the Create method
func (suite *RepositoryTestSuite) TestCreateStore() error {
	store := &domain.Store{
		Name:        "Test Store",
		FranchiseID: 1,
		IsActive:    true,
	}

	err := suite.StoreRepo.Create(store)
	if err != nil {
		return err
	}

	// Verify the store was created
	created, err := suite.StoreRepo.GetByID(int(store.ID))
	if err != nil {
		return err
	}

	if !suite.AssertStoreEqual(*store, created) {
		return errors.New("created store doesn't match expected")
	}

	return nil
}

// TestGetByFranchiseID demonstrates testing with specific conditions
func (suite *RepositoryTestSuite) TestGetByFranchiseID() error {
	suite.SeedTestData()

	stores, err := suite.StoreRepo.GetByFranchiseID(1)
	if err != nil {
		return err
	}

	if len(stores) != 2 {
		return errors.New("expected 2 stores for franchise 1")
	}

	return nil
}

// TestGetActiveStoresByFranchise tests filtering active stores by franchise
func (suite *RepositoryTestSuite) TestGetActiveStoresByFranchise() error {
	suite.SeedTestData()

	activeStores, err := suite.StoreRepo.GetActiveStoresByFranchise(1)
	if err != nil {
		return err
	}

	// Should return only active stores for franchise 1
	if len(activeStores) != 2 {
		return errors.New("expected 2 active stores for franchise 1")
	}

	for _, store := range activeStores {
		if !store.IsActive {
			return errors.New("returned store is not active")
		}
		if store.FranchiseID != 1 {
			return errors.New("returned store doesn't belong to franchise 1")
		}
	}

	return nil
}

// TestGetStoresWithEmployeeCount tests getting stores with employee counts
func (suite *RepositoryTestSuite) TestGetStoresWithEmployeeCount() error {
	suite.SeedTestData()

	storesWithCount, err := suite.StoreRepo.GetStoresWithEmployeeCount(1)
	if err != nil {
		return err
	}

	if len(storesWithCount) != 2 {
		return errors.New("expected 2 stores with employee count for franchise 1")
	}

	for _, storeWithCount := range storesWithCount {
		if storeWithCount.EmployeeCount <= 0 {
			return errors.New("employee count should be greater than 0")
		}
	}

	return nil
}

// TestGetStoreStatistics tests getting store statistics
func (suite *RepositoryTestSuite) TestGetStoreStatistics() error {
	suite.SeedTestData()

	stats, err := suite.StoreRepo.GetStoreStatistics(1)
	if err != nil {
		return err
	}

	if stats.EmployeeCount <= 0 {
		return errors.New("employee count should be greater than 0")
	}

	if stats.ShiftCount <= 0 {
		return errors.New("shift count should be greater than 0")
	}

	if stats.Store.ID != 1 {
		return errors.New("store statistics should be for store ID 1")
	}

	return nil
}

// TestCreateShift tests creating a shift
func (suite *RepositoryTestSuite) TestCreateShift() error {
	shift := domain.Shift{
		Name:      "Test Shift",
		StoreID:   1,
		StartTime: "09:00",
		EndTime:   "17:00",
		IsActive:  true,
	}

	err := suite.ShiftRepo.Create(shift)
	if err != nil {
		return err
	}

	// Verify the shift was created
	created, err := suite.ShiftRepo.GetByID(int(shift.ID))
	if err != nil {
		return err
	}

	if created.Name != shift.Name {
		return errors.New("created shift doesn't match expected")
	}

	return nil
}

// TestGetActiveShiftsByStore tests filtering active shifts by store
func (suite *RepositoryTestSuite) TestGetActiveShiftsByStore() error {
	suite.SeedTestData()

	activeShifts, err := suite.ShiftRepo.GetActiveShiftsByStore(1)
	if err != nil {
		return err
	}

	// Should return only active shifts for store 1
	if len(activeShifts) != 2 {
		return errors.New("expected 2 active shifts for store 1")
	}

	for _, shift := range activeShifts {
		if !shift.IsActive {
			return errors.New("returned shift is not active")
		}
		if shift.StoreID != 1 {
			return errors.New("returned shift doesn't belong to store 1")
		}
	}

	return nil
}

// TestGetShiftStatistics tests getting shift statistics
func (suite *RepositoryTestSuite) TestGetShiftStatistics() error {
	suite.SeedTestData()

	stats, err := suite.ShiftRepo.GetShiftStatistics(1)
	if err != nil {
		return err
	}

	if stats.TotalShifts != 3 {
		return errors.New("expected 3 total shifts for store 1")
	}

	if stats.ActiveShifts != 2 {
		return errors.New("expected 2 active shifts for store 1")
	}

	return nil
}

// MockShiftRepository implements repository.ShiftRepository for testing
type MockShiftRepository struct {
	shifts     map[uint]domain.Shift
	nextID     uint
	shouldFail bool
}

// Update implements repository.ShiftRepository.
func (m *MockShiftRepository) Update(shift domain.Shift) error {
	panic("unimplemented")
}

// NewMockShiftRepository creates a new mock shift repository
func NewMockShiftRepository() *MockShiftRepository {
	return &MockShiftRepository{
		shifts: make(map[uint]domain.Shift),
		nextID: 1,
	}
}

// SetShouldFail configures the mock to simulate failures
func (m *MockShiftRepository) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// SeedShiftData adds test shift data to the mock
func (m *MockShiftRepository) SeedShiftData(shifts []domain.Shift) {
	for _, shift := range shifts {
		if shift.ID == 0 {
			shift.ID = m.nextID
			m.nextID++
		}
		m.shifts[shift.ID] = shift
	}
}

// Create creates a new shift
func (m *MockShiftRepository) Create(shift domain.Shift) error {
	if m.shouldFail {
		return errors.New("mock error: Create failed")
	}

	if shift.Name == "" {
		return mysql.ErrInvalidInput
	}

	shift.ID = m.nextID
	m.nextID++
	m.shifts[shift.ID] = shift
	return nil
}

// ListAll returns all shifts
func (m *MockShiftRepository) ListAll() ([]domain.Shift, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: ListAll failed")
	}

	shifts := make([]domain.Shift, 0, len(m.shifts))
	for _, shift := range m.shifts {
		shifts = append(shifts, shift)
	}
	return shifts, nil
}

// ListByStore returns shifts by store ID
func (m *MockShiftRepository) ListByStore(storeID int) ([]domain.Shift, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: ListByStore failed")
	}

	var shifts []domain.Shift
	for _, shift := range m.shifts {
		if shift.StoreID == storeID {
			shifts = append(shifts, shift)
		}
	}
	return shifts, nil
}

// GetByID returns a shift by ID
func (m *MockShiftRepository) GetByID(id int) (*domain.Shift, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetByID failed")
	}

	shift, exists := m.shifts[uint(id)]
	if !exists {
		return nil, mysql.ErrNotFound
	}
	return &shift, nil
}

// GetActiveShiftsByStore returns only active shifts for a store
func (m *MockShiftRepository) GetActiveShiftsByStore(storeID int) ([]domain.Shift, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetActiveShiftsByStore failed")
	}

	var shifts []domain.Shift
	for _, shift := range m.shifts {
		if shift.StoreID == storeID && shift.IsActive {
			shifts = append(shifts, shift)
		}
	}
	return shifts, nil
}

// Delete removes a shift by ID (soft delete)
func (m *MockShiftRepository) Delete(id int) error {
	if m.shouldFail {
		return errors.New("mock error: Delete failed")
	}

	if id <= 0 {
		return errors.New("invalid shift ID")
	}

	// Find and soft delete the shift
	for i, shift := range m.shifts {
		if int(shift.ID) == id {
			if !shift.IsActive {
				return errors.New("shift is already inactive/deleted")
			}
			// Perform soft delete by setting IsActive to false
			shift.IsActive = false
			m.shifts[i] = shift
			return nil
		}
	}

	return errors.New("shift not found")
}

// GetShiftStatistics returns comprehensive shift statistics for a store
func (m *MockShiftRepository) GetShiftStatistics(storeID int) (*repository.ShiftStatistics, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetShiftStatistics failed")
	}

	// Count shifts for the store
	var totalShifts, activeShifts int64

	for _, shift := range m.shifts {
		if shift.StoreID == storeID {
			totalShifts++
			if shift.IsActive {
				activeShifts++
			}
		}
	}

	stats := &repository.ShiftStatistics{
		TotalShifts:  totalShifts,
		ActiveShifts: activeShifts,
	}

	return stats, nil
}

// MockAbsenceRepository implements repository.AbsenceRepository for testing
type MockAbsenceRepository struct {
	absences   map[uint]domain.Absence
	nextID     uint
	shouldFail bool
}

// NewMockAbsenceRepository creates a new mock absence repository
func NewMockAbsenceRepository() *MockAbsenceRepository {
	return &MockAbsenceRepository{
		absences: make(map[uint]domain.Absence),
		nextID:   1,
	}
}

// SetShouldFail configures the mock to simulate failures
func (m *MockAbsenceRepository) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// SeedAbsenceData adds test absence data to the mock
func (m *MockAbsenceRepository) SeedAbsenceData(absences []domain.Absence) {
	for _, absence := range absences {
		if absence.ID == 0 {
			absence.ID = m.nextID
			m.nextID++
		}
		m.absences[absence.ID] = absence
	}
}

// Create creates a new absence
func (m *MockAbsenceRepository) Create(absence *domain.Absence) error {
	if m.shouldFail {
		return errors.New("mock error: Create failed")
	}

	if absence.EmployeeID <= 0 {
		return mysql.ErrInvalidInput
	}

	absence.ID = m.nextID
	m.nextID++
	m.absences[absence.ID] = *absence
	return nil
}

// GetByEmployeeAndMonth returns absences by employee and month
func (m *MockAbsenceRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetByEmployeeAndMonth failed")
	}

	var absences []domain.Absence
	for _, absence := range m.absences {
		if absence.EmployeeID == employeeID &&
			absence.Date.Year() == year &&
			int(absence.Date.Month()) == month {
			absences = append(absences, absence)
		}
	}
	return absences, nil
}

// GetByEmployeeAndDateRange returns absences by employee and date range
func (m *MockAbsenceRepository) GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Absence, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetByEmployeeAndDateRange failed")
	}

	var absences []domain.Absence
	for _, absence := range m.absences {
		if absence.EmployeeID == employeeID &&
			!absence.Date.Before(from) &&
			!absence.Date.After(to) {
			absences = append(absences, absence)
		}
	}
	return absences, nil
}

// GetTotalHoursByEmployee returns total hours by employee and month
func (m *MockAbsenceRepository) GetTotalHoursByEmployee(employeeID, year, month int) (float64, error) {
	if m.shouldFail {
		return 0, errors.New("mock error: GetTotalHoursByEmployee failed")
	}

	var total float64
	for _, absence := range m.absences {
		if absence.EmployeeID == employeeID &&
			absence.Date.Year() == year &&
			int(absence.Date.Month()) == month {
			total += absence.Hours
		}
	}
	return total, nil
}

// MockNoveltyRepository implements repository.NoveltyRepository for testing
type MockNoveltyRepository struct {
	novelties  map[uint]domain.Novelty
	nextID     uint
	shouldFail bool
}

// NewMockNoveltyRepository creates a new mock novelty repository
func NewMockNoveltyRepository() *MockNoveltyRepository {
	return &MockNoveltyRepository{
		novelties: make(map[uint]domain.Novelty),
		nextID:    1,
	}
}

// SetShouldFail configures the mock to simulate failures
func (m *MockNoveltyRepository) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// SeedNoveltyData adds test novelty data to the mock
func (m *MockNoveltyRepository) SeedNoveltyData(novelties []domain.Novelty) {
	for _, novelty := range novelties {
		if novelty.ID == 0 {
			novelty.ID = m.nextID
			m.nextID++
		}
		m.novelties[novelty.ID] = novelty
	}
}

// Create creates a new novelty
func (m *MockNoveltyRepository) Create(novelty *domain.Novelty) error {
	if m.shouldFail {
		return errors.New("mock error: Create failed")
	}

	if novelty.EmployeeID <= 0 {
		return mysql.ErrInvalidInput
	}

	novelty.ID = m.nextID
	m.nextID++
	m.novelties[novelty.ID] = *novelty
	return nil
}

// GetByEmployeeAndMonth returns novelties by employee and month
func (m *MockNoveltyRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetByEmployeeAndMonth failed")
	}

	var novelties []domain.Novelty
	for _, novelty := range m.novelties {
		if novelty.EmployeeID == employeeID &&
			novelty.Date.Year() == year &&
			int(novelty.Date.Month()) == month {
			novelties = append(novelties, novelty)
		}
	}
	return novelties, nil
}

// GetByEmployeeAndDateRange returns novelties by employee and date range
func (m *MockNoveltyRepository) GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Novelty, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetByEmployeeAndDateRange failed")
	}

	var novelties []domain.Novelty
	for _, novelty := range m.novelties {
		if novelty.EmployeeID == employeeID &&
			!novelty.Date.Before(from) &&
			!novelty.Date.After(to) {
			novelties = append(novelties, novelty)
		}
	}
	return novelties, nil
}

// GetTotalHoursByEmployeeAndType returns total hours by employee, month, and type
func (m *MockNoveltyRepository) GetTotalHoursByEmployeeAndType(employeeID, year, month int, noveltyType string) (float64, error) {
	if m.shouldFail {
		return 0, errors.New("mock error: GetTotalHoursByEmployeeAndType failed")
	}

	var total float64
	for _, novelty := range m.novelties {
		if novelty.EmployeeID == employeeID &&
			novelty.Date.Year() == year &&
			int(novelty.Date.Month()) == month &&
			novelty.Type == noveltyType {
			total += novelty.Hours
		}
	}
	return total, nil
}

// GetNoveltyTypesSummary returns summary of novelty types for an employee and month
func (m *MockNoveltyRepository) GetNoveltyTypesSummary(employeeID, year, month int) ([]repository.NoveltyTypeSummary, error) {
	if m.shouldFail {
		return nil, errors.New("mock error: GetNoveltyTypesSummary failed")
	}

	// Count novelties by type
	typeCounts := make(map[string]int)
	typeHours := make(map[string]float64)

	for _, novelty := range m.novelties {
		if novelty.EmployeeID == employeeID &&
			novelty.Date.Year() == year &&
			int(novelty.Date.Month()) == month {
			typeCounts[novelty.Type]++
			typeHours[novelty.Type] += novelty.Hours
		}
	}

	// Convert to summary slice
	var summary []repository.NoveltyTypeSummary
	for noveltyType, count := range typeCounts {
		summary = append(summary, repository.NoveltyTypeSummary{
			Type:       noveltyType,
			Count:      count,
			TotalHours: typeHours[noveltyType],
		})
	}

	return summary, nil
}
