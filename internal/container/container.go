package container

import (
	"loopi-api/internal/delivery/http"
	"loopi-api/internal/repository"
	mysqlRepo "loopi-api/internal/repository/mysql"
	"loopi-api/internal/usecase"

	"gorm.io/gorm"
)

// Container holds all dependencies for the application
type Container struct {
	// Database
	DB *gorm.DB

	// Repositories
	Repositories *Repositories

	// UseCases
	UseCases *UseCases

	// Handlers
	Handlers *Handlers
}

// Repositories contains all repository implementations
type Repositories struct {
	User          repository.UserRepository
	Franchise     repository.FranchiseRepository
	Store         repository.StoreRepository
	AssignedShift repository.AssignedShiftRepository
	Absence       repository.AbsenceRepository
	Novelty       repository.NoveltyRepository
	Shift         repository.ShiftRepository
	WorkConfig    repository.WorkConfigRepository
}

// UseCases contains all use case implementations
type UseCases struct {
	Auth            usecase.AuthUseCase
	Franchise       usecase.FranchiseUseCase
	Store           usecase.StoreUseCase
	Employee        usecase.EmployeeUseCase
	EmployeeHours   usecase.EmployeeHoursUseCase
	Shift           usecase.ShiftUseCase
	ShiftProjection usecase.ShiftProjectionUseCase
	Calendar        usecase.CalendarUseCase
	Absence         usecase.AbsenceUseCase
	Novelty         usecase.NoveltyUseCase
}

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth            *http.AuthHandler
	Franchise       *http.FranchiseHandler
	Store           *http.StoreHandler
	Employee        *http.EmployeeHandler
	EmployeeHours   *http.EmployeeHoursHandler
	Shift           *http.ShiftHandler
	ShiftProjection *http.ShiftProjectionHandler
	Calendar        *http.CalendarHandler
	Absence         *http.AbsenceHandler
	Novelty         *http.NoveltyHandler
}

// NewContainer creates a new dependency container
func NewContainer(db *gorm.DB) *Container {
	container := &Container{
		DB: db,
	}

	// Initialize repositories
	container.Repositories = newRepositories(db)

	// Initialize use cases
	container.UseCases = newUseCases(container.Repositories)

	// Initialize handlers
	container.Handlers = newHandlers(container.UseCases)

	return container
}

// newRepositories creates all repository instances
func newRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:          mysqlRepo.NewUserRepository(db),
		Franchise:     mysqlRepo.NewFranchiseRepository(db),
		Store:         mysqlRepo.NewStoreRepository(db),
		AssignedShift: mysqlRepo.NewAssignedShiftRepository(db),
		Absence:       mysqlRepo.NewAbsenceRepository(db),
		Novelty:       mysqlRepo.NewNoveltyRepository(db),
		Shift:         mysqlRepo.NewShiftRepository(db),
		WorkConfig:    mysqlRepo.NewWorkConfigRepository(db),
	}
}

// newUseCases creates all use case instances
func newUseCases(repos *Repositories) *UseCases {
	return &UseCases{
		Auth:            usecase.NewAuthUseCase(repos.User),
		Franchise:       usecase.NewFranchiseUseCase(repos.Franchise),
		Store:           usecase.NewStoreUseCase(repos.Store),
		Employee:        usecase.NewEmployeeUseCase(repos.User),
		EmployeeHours:   usecase.NewEmployeeHoursUseCase(repos.AssignedShift, repos.Absence, repos.Novelty, repos.User),
		Shift:           usecase.NewShiftUseCase(repos.Shift),
		ShiftProjection: usecase.NewShiftProjectionUseCase(repos.Shift, repos.WorkConfig),
		Calendar:        usecase.NewCalendarUseCase(),
		Absence:         usecase.NewAbsenceUseCase(repos.Absence),
		Novelty:         usecase.NewNoveltyUseCase(repos.Novelty),
	}
}

// newHandlers creates all HTTP handler instances
func newHandlers(useCases *UseCases) *Handlers {
	return &Handlers{
		Auth:            http.NewAuthHandler(useCases.Auth),
		Franchise:       http.NewFranchiseHandler(useCases.Franchise),
		Store:           http.NewStoreHandler(useCases.Store),
		Employee:        http.NewEmployeeHandler(useCases.Employee),
		EmployeeHours:   http.NewEmployeeHoursHandler(useCases.EmployeeHours),
		Shift:           http.NewShiftHandler(useCases.Shift),
		ShiftProjection: http.NewShiftProjectionHandler(useCases.ShiftProjection),
		Calendar:        http.NewCalendarHandler(useCases.Calendar),
		Absence:         http.NewAbsenceHandler(useCases.Absence),
		Novelty:         http.NewNoveltyHandler(useCases.Novelty),
	}
}
