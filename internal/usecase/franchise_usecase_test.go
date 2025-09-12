package usecase

import (
	"errors"
	"testing"

	"loopi-api/internal/domain"
)

// MockFranchiseRepository for testing
type MockFranchiseRepository struct {
	franchises   []domain.Franchise
	shouldError  bool
	errorMessage string
}

func (m *MockFranchiseRepository) GetAll() ([]domain.Franchise, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMessage)
	}
	return m.franchises, nil
}

func (m *MockFranchiseRepository) GetById(id int) (domain.Franchise, error) {
	if m.shouldError {
		return domain.Franchise{}, errors.New(m.errorMessage)
	}

	for _, franchise := range m.franchises {
		if int(franchise.ID) == id {
			return franchise, nil
		}
	}
	return domain.Franchise{}, errors.New("franchise not found")
}

func (m *MockFranchiseRepository) Create(franchise *domain.Franchise) error {
	if m.shouldError {
		return errors.New(m.errorMessage)
	}

	// Simulate ID assignment
	franchise.ID = uint(len(m.franchises) + 1)
	m.franchises = append(m.franchises, *franchise)
	return nil
}

// Helper function to create test franchises
func createTestFranchises() []domain.Franchise {
	return []domain.Franchise{
		{
			BaseEntity: domain.BaseEntity{ID: 1},
			Name:       "Franchise A",
			IsActive:   true,
		},
		{
			BaseEntity: domain.BaseEntity{ID: 2},
			Name:       "Franchise B",
			IsActive:   true,
		},
		{
			BaseEntity: domain.BaseEntity{ID: 3},
			Name:       "Franchise C Inactive",
			IsActive:   false,
		},
	}
}

func TestFranchiseUseCase_GetAll_Success(t *testing.T) {
	// Arrange
	testFranchises := createTestFranchises()
	mockRepo := &MockFranchiseRepository{
		franchises:  testFranchises,
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	// Act
	franchises, err := useCase.GetAll()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(franchises) != 3 {
		t.Errorf("Expected 3 franchises, got %d", len(franchises))
	}

	if franchises[0].Name != "Franchise A" {
		t.Errorf("Expected first franchise name to be 'Franchise A', got %s", franchises[0].Name)
	}
}

func TestFranchiseUseCase_GetAll_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockFranchiseRepository{
		shouldError:  true,
		errorMessage: "database connection failed",
	}
	useCase := NewFranchiseUseCase(mockRepo)

	// Act
	franchises, err := useCase.GetAll()

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if franchises != nil {
		t.Error("Expected nil franchises on error")
	}

	// Check that error is properly wrapped
	if err.Error() == "database connection failed" {
		t.Error("Error should be wrapped with context, not raw repository error")
	}
}

func TestFranchiseUseCase_GetAll_NoFranchises(t *testing.T) {
	// Arrange
	mockRepo := &MockFranchiseRepository{
		franchises:  []domain.Franchise{}, // Empty slice
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	// Act
	franchises, err := useCase.GetAll()

	// Assert
	if err == nil {
		t.Error("Expected error for no franchises found, got nil")
	}

	if franchises != nil {
		t.Error("Expected nil franchises when none found")
	}
}

func TestFranchiseUseCase_GetById_Success(t *testing.T) {
	// Arrange
	testFranchises := createTestFranchises()
	mockRepo := &MockFranchiseRepository{
		franchises:  testFranchises,
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	// Act
	franchise, err := useCase.GetById(1)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if franchise.ID != 1 {
		t.Errorf("Expected franchise ID 1, got %d", franchise.ID)
	}

	if franchise.Name != "Franchise A" {
		t.Errorf("Expected franchise name 'Franchise A', got %s", franchise.Name)
	}
}

func TestFranchiseUseCase_GetById_InvalidID(t *testing.T) {
	// Arrange
	testFranchises := createTestFranchises()
	mockRepo := &MockFranchiseRepository{
		franchises:  testFranchises,
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	// Act
	franchise, err := useCase.GetById(-1) // Invalid ID

	// Assert
	if err == nil {
		t.Error("Expected validation error for invalid ID, got nil")
	}

	if franchise.ID != 0 {
		t.Error("Expected zero-value franchise on validation error")
	}
}

func TestFranchiseUseCase_GetById_NotFound(t *testing.T) {
	// Arrange
	testFranchises := createTestFranchises()
	mockRepo := &MockFranchiseRepository{
		franchises:  testFranchises,
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	// Act
	franchise, err := useCase.GetById(999) // Non-existent ID

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent franchise, got nil")
	}

	if franchise.ID != 0 {
		t.Error("Expected zero-value franchise on not found error")
	}
}

func TestFranchiseUseCase_Create_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockFranchiseRepository{
		franchises:  []domain.Franchise{},
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	newFranchise := domain.Franchise{
		Name: "New Franchise",
	}

	// Act
	err := useCase.Create(newFranchise)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that franchise was added to repository
	if len(mockRepo.franchises) != 1 {
		t.Errorf("Expected 1 franchise in repository, got %d", len(mockRepo.franchises))
	}

	createdFranchise := mockRepo.franchises[0]
	if !createdFranchise.IsActive {
		t.Error("Expected created franchise to be active by default")
	}

	if createdFranchise.Name != "New Franchise" {
		t.Errorf("Expected franchise name 'New Franchise', got %s", createdFranchise.Name)
	}
}

func TestFranchiseUseCase_Create_ValidationError(t *testing.T) {
	// Arrange
	mockRepo := &MockFranchiseRepository{
		franchises:  []domain.Franchise{},
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	invalidFranchise := domain.Franchise{
		Name: "", // Empty name should fail validation
	}

	// Act
	err := useCase.Create(invalidFranchise)

	// Assert
	if err == nil {
		t.Error("Expected validation error for empty name, got nil")
	}

	// Check that no franchise was added to repository
	if len(mockRepo.franchises) != 0 {
		t.Error("Expected no franchise to be added on validation error")
	}
}

func TestFranchiseUseCase_GetActiveFranchises_Success(t *testing.T) {
	// Arrange
	testFranchises := createTestFranchises()
	mockRepo := &MockFranchiseRepository{
		franchises:  testFranchises,
		shouldError: false,
	}
	useCase := NewFranchiseUseCase(mockRepo)

	// Act
	activeFranchises, err := useCase.GetActiveFranchises()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should only return active franchises (2 out of 3)
	if len(activeFranchises) != 2 {
		t.Errorf("Expected 2 active franchises, got %d", len(activeFranchises))
	}

	// Verify all returned franchises are active
	for _, franchise := range activeFranchises {
		if !franchise.IsActive {
			t.Errorf("Found inactive franchise in active franchises list: %s", franchise.Name)
		}
	}
}

func TestFranchiseUseCase_ValidateFranchiseData_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockFranchiseRepository{}
	useCase := NewFranchiseUseCase(mockRepo)

	validFranchise := &domain.Franchise{
		Name: "Valid Franchise Name",
	}

	// Act
	err := useCase.ValidateFranchiseData(validFranchise)

	// Assert
	if err != nil {
		t.Errorf("Expected no validation error, got %v", err)
	}
}

func TestFranchiseUseCase_ValidateFranchiseData_FailsOnEmptyName(t *testing.T) {
	// Arrange
	mockRepo := &MockFranchiseRepository{}
	useCase := NewFranchiseUseCase(mockRepo)

	invalidFranchise := &domain.Franchise{
		Name: "", // Empty name
	}

	// Act
	err := useCase.ValidateFranchiseData(invalidFranchise)

	// Assert
	if err == nil {
		t.Error("Expected validation error for empty name, got nil")
	}
}

func TestFranchiseUseCase_ValidateFranchiseData_FailsOnShortName(t *testing.T) {
	// Arrange
	mockRepo := &MockFranchiseRepository{}
	useCase := NewFranchiseUseCase(mockRepo)

	invalidFranchise := &domain.Franchise{
		Name: "AB", // Too short (less than 3 characters)
	}

	// Act
	err := useCase.ValidateFranchiseData(invalidFranchise)

	// Assert
	if err == nil {
		t.Error("Expected validation error for short name, got nil")
	}
}
