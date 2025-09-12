package base

// EntityWithID interface for entities that have an ID
type EntityWithID interface {
	GetID() uint
}

// BaseRepository interface for the most basic operations
type BaseRepository[T EntityWithID] interface {
	GetAll() ([]T, error)
	Create(entity *T) error
}

// ExtendedRepository interface for repositories with full CRUD
type ExtendedRepository[T EntityWithID] interface {
	BaseRepository[T]
	GetByID(id int) (T, error)
	Update(entity *T) error
	Delete(id int) error
}

// RepositoryAdapter adapts existing repositories to work with BaseUseCase
type RepositoryAdapter[T EntityWithID] struct {
	GetAllFunc  func() ([]T, error)
	GetByIDFunc func(id int) (T, error)
	CreateFunc  func(entity *T) error
	UpdateFunc  func(entity *T) error
	DeleteFunc  func(id int) error
}

func (ra *RepositoryAdapter[T]) GetAll() ([]T, error) {
	if ra.GetAllFunc != nil {
		return ra.GetAllFunc()
	}
	var zero []T
	return zero, nil
}

func (ra *RepositoryAdapter[T]) GetByID(id int) (T, error) {
	var zero T
	if ra.GetByIDFunc != nil {
		return ra.GetByIDFunc(id)
	}
	return zero, nil
}

func (ra *RepositoryAdapter[T]) Create(entity *T) error {
	if ra.CreateFunc != nil {
		return ra.CreateFunc(entity)
	}
	return nil
}

func (ra *RepositoryAdapter[T]) Update(entity *T) error {
	if ra.UpdateFunc != nil {
		return ra.UpdateFunc(entity)
	}
	return nil
}

func (ra *RepositoryAdapter[T]) Delete(id int) error {
	if ra.DeleteFunc != nil {
		return ra.DeleteFunc(id)
	}
	return nil
}

// BaseUseCase provides common CRUD operations for use cases
type BaseUseCase[T EntityWithID] struct {
	repo         *RepositoryAdapter[T]
	entityName   string
	errorHandler *ErrorHandler
	validator    *Validator
	logger       *Logger
}

// NewBaseUseCase creates a new base use case with adapter
func NewBaseUseCase[T EntityWithID](entityName string) *BaseUseCase[T] {
	return &BaseUseCase[T]{
		repo:         &RepositoryAdapter[T]{},
		entityName:   entityName,
		errorHandler: NewErrorHandler(entityName),
		validator:    NewValidator(),
		logger:       NewLogger(entityName),
	}
}

// SetRepository sets the repository functions
func (uc *BaseUseCase[T]) SetRepository(adapter *RepositoryAdapter[T]) {
	uc.repo = adapter
}

// GetAll retrieves all entities with proper error handling
func (uc *BaseUseCase[T]) GetAll() ([]T, error) {
	uc.logger.LogOperation("GetAll", "start", nil)

	entities, err := uc.repo.GetAll()
	if err != nil {
		uc.logger.LogError("GetAll", err, nil)
		return nil, uc.errorHandler.HandleRepositoryError("GetAll", err)
	}

	if len(entities) == 0 {
		uc.logger.LogOperation("GetAll", "no_entities_found", nil)
		return nil, uc.errorHandler.HandleNotFound("GetAll", "No entities found")
	}

	uc.logger.LogOperation("GetAll", "success", map[string]interface{}{
		"count": len(entities),
	})

	return entities, nil
}

// GetByID retrieves an entity by ID with proper error handling
func (uc *BaseUseCase[T]) GetByID(id int) (T, error) {
	var zero T

	uc.logger.LogOperation("GetByID", "start", map[string]interface{}{"id": id})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("GetByID", err, map[string]interface{}{"id": id})
		return zero, uc.errorHandler.HandleValidationError("GetByID", err)
	}

	entity, err := uc.repo.GetByID(id)
	if err != nil {
		uc.logger.LogError("GetByID", err, map[string]interface{}{"id": id})
		return zero, uc.errorHandler.HandleRepositoryError("GetByID", err)
	}

	uc.logger.LogOperation("GetByID", "success", map[string]interface{}{"id": id})
	return entity, nil
}

// Create creates a new entity with validation and error handling
func (uc *BaseUseCase[T]) Create(entity *T) error {
	uc.logger.LogOperation("Create", "start", map[string]interface{}{
		"entity_type": uc.entityName,
	})

	// Validate entity
	if err := uc.validator.ValidateEntity(entity); err != nil {
		uc.logger.LogError("Create", err, nil)
		return uc.errorHandler.HandleValidationError("Create", err)
	}

	// Execute creation
	if err := uc.repo.Create(entity); err != nil {
		uc.logger.LogError("Create", err, nil)
		return uc.errorHandler.HandleRepositoryError("Create", err)
	}

	uc.logger.LogOperation("Create", "success", map[string]interface{}{
		"entity_id": (*entity).GetID(),
	})

	return nil
}

// Update updates an existing entity with validation
func (uc *BaseUseCase[T]) Update(entity *T) error {
	entityID := (*entity).GetID()

	uc.logger.LogOperation("Update", "start", map[string]interface{}{
		"entity_id": entityID,
	})

	// Validate entity
	if err := uc.validator.ValidateEntity(entity); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{"entity_id": entityID})
		return uc.errorHandler.HandleValidationError("Update", err)
	}

	// Validate ID
	if err := uc.validator.ValidateID(int(entityID)); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{"entity_id": entityID})
		return uc.errorHandler.HandleValidationError("Update", err)
	}

	// Execute update
	if err := uc.repo.Update(entity); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{"entity_id": entityID})
		return uc.errorHandler.HandleRepositoryError("Update", err)
	}

	uc.logger.LogOperation("Update", "success", map[string]interface{}{
		"entity_id": entityID,
	})

	return nil
}

// Delete removes an entity by ID
func (uc *BaseUseCase[T]) Delete(id int) error {
	uc.logger.LogOperation("Delete", "start", map[string]interface{}{"id": id})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{"id": id})
		return uc.errorHandler.HandleValidationError("Delete", err)
	}

	// Execute deletion
	if err := uc.repo.Delete(id); err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{"id": id})
		return uc.errorHandler.HandleRepositoryError("Delete", err)
	}

	uc.logger.LogOperation("Delete", "success", map[string]interface{}{"id": id})
	return nil
}

// GetErrorHandler returns the error handler for custom operations
func (uc *BaseUseCase[T]) GetErrorHandler() *ErrorHandler {
	return uc.errorHandler
}

// GetValidator returns the validator for custom operations
func (uc *BaseUseCase[T]) GetValidator() *Validator {
	return uc.validator
}

// GetLogger returns the logger for custom operations
func (uc *BaseUseCase[T]) GetLogger() *Logger {
	return uc.logger
}

// GetRepository returns the repository adapter for custom operations
func (uc *BaseUseCase[T]) GetRepository() *RepositoryAdapter[T] {
	return uc.repo
}
