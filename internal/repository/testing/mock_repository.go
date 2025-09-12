package testing

import (
	"errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/repository/mysql"
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

// RepositoryTestSuite provides common test utilities for repositories
type RepositoryTestSuite struct {
	StoreRepo repository.StoreRepository
}

// NewRepositoryTestSuite creates a new test suite
func NewRepositoryTestSuite(useRealDB bool) *RepositoryTestSuite {
	var storeRepo repository.StoreRepository

	if useRealDB {
		// In real tests, you would inject a test database connection
		// storeRepo = mysql.NewStoreRepository(testDB)
		panic("Real DB testing not implemented in this example")
	} else {
		storeRepo = NewMockStoreRepository()
	}

	return &RepositoryTestSuite{
		StoreRepo: storeRepo,
	}
}

// SeedTestData adds common test data
func (suite *RepositoryTestSuite) SeedTestData() {
	if mockRepo, ok := suite.StoreRepo.(*MockStoreRepository); ok {
		testStores := []domain.Store{
			{Name: "Store 1", FranchiseID: 1, IsActive: true},
			{Name: "Store 2", FranchiseID: 1, IsActive: true},
			{Name: "Store 3", FranchiseID: 2, IsActive: false},
		}
		mockRepo.SeedData(testStores)
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
