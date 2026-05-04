# API Endpoints — Loopi API

**Fuente**: Análisis de código `internal/interface/router/router.go`
**Total**: 37 endpoints

## Rutas Públicas

| Método | Path | Handler | Auth |
|--------|------|---------|------|
| GET | `/health` | inline | Ninguna |
| POST | `/api/auth/login` | `authHandler.Login` | Ninguna |

## Rutas Protegidas (JWT requerido — cualquier rol)

| Método | Path | Handler | Descripción |
|--------|------|---------|-------------|
| GET | `/api/employees/me` | `authHandler.GetMe` | Perfil del empleado autenticado |
| GET | `/api/inventories/latest` | `employeeInventoryHandler.GetLatest` | Último inventario completado |
| GET | `/api/inventories/in-progress` | `employeeInventoryHandler.GetInProgress` | Inventarios en progreso del empleado |
| GET | `/api/inventories/suggested-schedule` | `employeeInventoryHandler.GetSuggestedSchedule` | Schedule sugerido según hora actual |
| POST | `/api/inventories` | `employeeInventoryHandler.Create` | Crear inventario |
| GET | `/api/inventories/{inventoryID}/items` | `employeeInventoryHandler.GetItems` | Items del inventario con valores sugeridos |
| POST | `/api/inventories/{inventoryID}/details` | `employeeInventoryHandler.SaveDetail` | Guardar conteo físico de un item |
| GET | `/api/inventories/{inventoryID}/discrepancies` | `employeeInventoryHandler.GetDiscrepancies` | Items con discrepancia |
| POST | `/api/inventories/{inventoryID}/sales` | `employeeInventoryHandler.SaveSales` | Guardar ventas y compras |
| GET | `/api/inventories/{inventoryID}/summary` | `employeeInventoryHandler.GetSummary` | Resumen antes de completar |
| POST | `/api/inventories/{inventoryID}/complete` | `employeeInventoryHandler.Complete` | Completar inventario |

## Rutas Admin (JWT + rol "admin")

| Método | Path | Handler | Descripción |
|--------|------|---------|-------------|
| GET | `/api/admin/dashboard` | `adminDashboardHandler.GetDashboard` | Dashboard con stats del día |
| GET | `/api/admin/inventories` | `adminInventoryHandler.ListInventories` | Lista paginada con filtros |
| GET | `/api/admin/inventories/active-count` | `adminInventoryHandler.GetActiveInventoriesCount` | Conteo de inventarios en progreso |
| POST | `/api/admin/inventories/initial` | `adminInventoryHandler.CreateInitialInventory` | Crear inventario inicial (baseline) |
| GET | `/api/admin/inventories/{inventoryID}` | `adminInventoryHandler.GetInventoryDetail` | Detalle completo |
| PUT | `/api/admin/inventories/{inventoryID}/details/{detailID}` | `adminInventoryHandler.UpdateInventoryDetail` | Editar detalle (incluso cerrado) |
| GET | `/api/admin/measurement-units` | `adminItemHandler.ListMeasurementUnits` | Unidades de medida |
| GET | `/api/admin/items` | `adminItemHandler.ListItems` | Lista paginada de items |
| POST | `/api/admin/items` | `adminItemHandler.CreateItem` | Crear item |
| GET | `/api/admin/items/{itemID}` | `adminItemHandler.GetItem` | Detalle de item |
| PUT | `/api/admin/items/{itemID}` | `adminItemHandler.UpdateItem` | Actualizar item |
| PATCH | `/api/admin/items/{itemID}/status` | `adminItemHandler.UpdateItemStatus` | Activar/desactivar item |
| GET | `/api/admin/employees` | `adminEmployeeHandler.ListEmployees` | Lista paginada de empleados |
| GET | `/api/admin/employees/active` | `adminEmployeeHandler.ListAllActiveEmployees` | Todos los activos |
| POST | `/api/admin/employees` | `adminEmployeeHandler.CreateEmployee` | Crear empleado |
| GET | `/api/admin/employees/{employeeID}` | `adminEmployeeHandler.GetEmployee` | Detalle de empleado |
| PUT | `/api/admin/employees/{employeeID}` | `adminEmployeeHandler.UpdateEmployee` | Actualizar empleado |
| PATCH | `/api/admin/employees/{employeeID}/status` | `adminEmployeeHandler.UpdateEmployeeStatus` | Activar/desactivar |
| POST | `/api/admin/employees/{employeeID}/reset-password` | `adminEmployeeHandler.ResetEmployeePassword` | Reset password |
| GET | `/api/admin/categories` | `adminCategoryHandler.ListCategories` | Listar categorías |
| POST | `/api/admin/categories` | `adminCategoryHandler.CreateCategory` | Crear categoría |
| POST | `/api/admin/categories/reorder` | `adminCategoryHandler.ReorderCategories` | Reordenar |
| GET | `/api/admin/categories/{categoryID}` | `adminCategoryHandler.GetCategory` | Detalle |
| PUT | `/api/admin/categories/{categoryID}` | `adminCategoryHandler.UpdateCategory` | Actualizar |
| PATCH | `/api/admin/categories/{categoryID}/status` | `adminCategoryHandler.UpdateCategoryStatus` | Activar/desactivar |
| GET | `/api/admin/suppliers` | `adminSupplierHandler.ListSuppliers` | Lista paginada |
| GET | `/api/admin/suppliers/active` | `adminSupplierHandler.ListAllActiveSuppliers` | Todos activos |
| POST | `/api/admin/suppliers` | `adminSupplierHandler.CreateSupplier` | Crear proveedor |
| GET | `/api/admin/suppliers/{supplierID}` | `adminSupplierHandler.GetSupplier` | Detalle |
| PUT | `/api/admin/suppliers/{supplierID}` | `adminSupplierHandler.UpdateSupplier` | Actualizar |
| PATCH | `/api/admin/suppliers/{supplierID}/status` | `adminSupplierHandler.UpdateSupplierStatus` | Activar/desactivar |
