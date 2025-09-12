# Repository Layer - Mejoras de Mantenibilidad

## 🎯 Objetivos

Esta refactorización del package repository introduce varias mejoras de mantenibilidad:

- **Reutilización de código**: Base repository con operaciones CRUD comunes
- **Manejo de errores estandarizado**: Errores consistentes y descriptivos  
- **Query helpers**: Utilidades para consultas comunes
- **Testing mejorado**: Mocks y herramientas para testing
- **Consistencia**: Patrones uniformes en todos los repositorios

## 📁 Estructura

```
internal/repository/
├── mysql/
│   ├── base_repository.go      # Repositorio base con CRUD genérico
│   ├── errors.go              # Manejo estandarizado de errores
│   ├── query_helpers.go       # Helpers para consultas comunes
│   ├── store_repository.go    # Implementación original
│   └── store_repository_refactored.go  # Ejemplo refactorizado
├── testing/
│   └── mock_repository.go     # Mocks para testing
└── README.md                  # Esta documentación
```

## 🔧 Cómo Usar

### 1. BaseRepository - Operaciones CRUD Comunes

```go
// Crear un repositorio básico
baseRepo := mysql.NewBaseRepository[domain.Store](db, "stores")

// Operaciones CRUD automáticas
store, err := baseRepo.GetByID(1)
stores, err := baseRepo.GetAll()
err = baseRepo.Create(&store)
err = baseRepo.Update(&store) 
err = baseRepo.Delete(1)

// Operaciones avanzadas
stores, err := baseRepo.FindBy(map[string]interface{}{
    "franchise_id": 1,
    "is_active": true,
})

count, err := baseRepo.Count()
exists, err := baseRepo.Exists(1)
```

### 2. Manejo de Errores Estandarizado

```go
// En tu repositorio
type storeRepository struct {
    *mysql.BaseRepository[domain.Store]
    errorHandler *mysql.ErrorHandler
}

func NewStoreRepository(db *gorm.DB) repository.StoreRepository {
    return &storeRepository{
        BaseRepository: mysql.NewBaseRepository[domain.Store](db, "stores"),
        errorHandler:   mysql.NewErrorHandler("stores"),
    }
}

// Uso del error handler
func (r *storeRepository) GetByID(id int) (domain.Store, error) {
    store, err := r.BaseRepository.GetByID(id)
    if err != nil {
        if err == mysql.ErrNotFound {
            return domain.Store{}, r.errorHandler.HandleNotFound("GetByID", id)
        }
        return domain.Store{}, r.errorHandler.HandleError("GetByID", err, id)
    }
    return *store, nil
}
```

### 3. Query Helpers

```go
// Query builder fluente
stores, err := mysql.NewQueryBuilder(db).
    WhereEquals("franchise_id", 1).
    WhereActive().
    OrderBy("name").
    Limit(10).
    GetDB().
    Find(&stores)

// Patrones comunes predefinidos
stores, err := mysql.FindActiveByFranchise[domain.Store](db, franchiseID)
absences, err := mysql.FindByEmployeeAndMonth[domain.Absence](db, empID, year, month)

// Paginación
pagination := mysql.NewPaginationHelper(page, pageSize)
stores, err := mysql.FindWithPagination[domain.Store](db, pagination, conditions)
```

### 4. Testing con Mocks

```go
// Setup de test
func TestStoreService(t *testing.T) {
    // Usar mock repository
    suite := testing.NewRepositoryTestSuite(false) // false = use mocks
    suite.SeedTestData()
    
    // Tu servicio que usa el repositorio
    service := usecase.NewStoreUseCase(suite.StoreRepo)
    
    // Test normal
    store, err := service.GetByID(1)
    assert.NoError(t, err)
    assert.Equal(t, "Store 1", store.Name)
    
    // Test error scenarios
    if mockRepo, ok := suite.StoreRepo.(*testing.MockStoreRepository); ok {
        mockRepo.SetShouldFail(true)
        _, err = service.GetByID(1)
        assert.Error(t, err)
    }
}
```

## 🔄 Migración Gradual

### Paso 1: Mantener compatibilidad
Los repositorios existentes siguen funcionando sin cambios.

### Paso 2: Refactorizar uno por uno
```go
// Antes
type storeRepo struct {
    db *gorm.DB
}

func (r *storeRepo) GetAll() ([]domain.Store, error) {
    var stores []domain.Store
    err := r.db.Find(&stores).Error
    return stores, err
}

// Después  
type storeRepository struct {
    *mysql.BaseRepository[domain.Store]
    errorHandler *mysql.ErrorHandler
}

func (r *storeRepository) GetAll() ([]domain.Store, error) {
    stores, err := r.BaseRepository.GetAll()
    if err != nil {
        return nil, r.errorHandler.HandleError("GetAll", err)
    }
    return stores, nil
}
```

### Paso 3: Añadir funcionalidad específica
```go
// Métodos específicos del dominio usando los helpers
func (r *storeRepository) GetActiveStoresByFranchise(franchiseID int) ([]domain.Store, error) {
    var stores []domain.Store
    err := mysql.NewQueryBuilder(r.GetDB()).
        WhereEquals("franchise_id", franchiseID).
        WhereActive().
        OrderBy("name").
        GetDB().
        Find(&stores).Error

    if err != nil {
        return nil, r.errorHandler.HandleError("GetActiveStoresByFranchise", err, franchiseID)
    }
    return stores, nil
}
```

## ✅ Beneficios

### 1. **Menos Código Duplicado**
- Operaciones CRUD comunes en BaseRepository
- Manejo de errores centralizado
- Query patterns reutilizables

### 2. **Errores Más Descriptivos**
```go
// Antes
error: record not found

// Después  
repository error: GetByID operation failed on table stores for ID 123: record not found
```

### 3. **Testing Más Fácil**
- Mocks automáticos
- Test utilities comunes
- Separación clara de concerns

### 4. **Consistencia**
- Mismo patrón en todos los repos
- Nomenclatura estandarizada
- Manejo uniforme de errores

### 5. **Escalabilidad**
- Fácil añadir nuevos repositorios
- Extensible con nuevas funcionalidades
- Mantiene la flexibilidad

## 🧪 Ejemplos de Uso en Producción

```go
// En tu container de dependencias
func newRepositories(db *gorm.DB) *Repositories {
    return &Repositories{
        Store:     mysql.NewStoreRepositoryRefactored(db),
        User:      mysql.NewUserRepositoryRefactored(db),
        Franchise: mysql.NewFranchiseRepositoryRefactored(db),
        // ... otros repos
    }
}

// En tus tests
func TestStoreUseCase(t *testing.T) {
    suite := testing.NewRepositoryTestSuite(false)
    useCase := usecase.NewStoreUseCase(suite.StoreRepo)
    
    // Tests con datos predefinidos
    suite.SeedTestData()
    stores, err := useCase.GetByFranchiseID(1)
    assert.NoError(t, err)
    assert.Len(t, stores, 2)
}
```

## 🎯 Próximos Pasos

1. **Implementar en repositorios restantes** usando el patrón mostrado
2. **Añadir logging** al ErrorHandler para debugging
3. **Extender QueryBuilder** con más operaciones según necesidades
4. **Implementar caching layer** en BaseRepository si es necesario
5. **Añadir métricas** para monitoreo de performance

Esta estructura mantiene la flexibilidad mientras reduce significativamente el código duplicado y mejora la mantenibilidad del proyecto.
