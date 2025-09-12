# 📬 Colecciones de Postman - Loopi API

Este documento te guía sobre cómo importar y usar las colecciones de Postman para probar todos los endpoints de la API Loopi.

## 🚀 Importar la Colección

1. **Abrir Postman**: Ejecuta la aplicación Postman en tu sistema
2. **Importar archivo**:
   - Clic en "Import" en la esquina superior izquierda
   - Selecciona "Upload Files"
   - Busca y selecciona el archivo `postman_collections.json`
   - Clic en "Import"

## 🔧 Configuración Inicial

### Variables de Entorno

La colección incluye variables predefinidas que puedes modificar según tu entorno:

- `baseUrl`: URL base de la API (por defecto: `http://localhost:8080`)
- `token`: Token JWT (se llena automáticamente al hacer login)
- `franchiseId`: ID de franquicia (por defecto: `1`)
- `storeId`: ID de tienda (por defecto: `1`)
- `employeeId`: ID de empleado (por defecto: `1`)
- `shiftId`: ID de turno (por defecto: `1`)

### 🔐 Autenticación

1. **Ejecutar Login**: Ir a `🔐 Authentication > Login`
2. **Modificar credenciales**: Edita el body con credenciales válidas
3. **Ejecutar**: El token se guardará automáticamente en la variable `token`
4. **Seleccionar contexto**: Ejecuta `🔐 Authentication > Select Context` si necesitas cambiar franquicia/tienda

## 📋 Estructura de las Colecciones

### 🔐 Authentication

- **Login**: Autenticación de usuario
- **Select Context**: Selección de contexto de franquicia/tienda

### 🏢 Franchises

- **Get All Franchises**: Obtener todas las franquicias
- **Get Franchise by ID**: Obtener franquicia específica
- **Create Franchise**: Crear nueva franquicia

### 🏪 Stores

- **CRUD Operations**: Crear, leer, actualizar, eliminar tiendas
- **Get by Franchise**: Obtener tiendas por franquicia
- **Statistics**: Estadísticas de tiendas

### 👥 Employees

- **CRUD Operations**: Gestión completa de empleados
- **Filter by Store**: Empleados por tienda
- **Active Employees**: Solo empleados activos

### ⏰ Employee Hours

- **Monthly Summary**: Resumen mensual de horas
- **Daily Summary**: Resumen diario de horas
- **Yearly Summary**: Resumen anual de horas
- **Working Days**: Días trabajados

### 📅 Calendar

- **Holidays**: Gestión de feriados
- **Month Summary**: Resumen mensual del calendario
- **Working Days**: Días laborables
- **Clear Cache**: Limpiar caché del calendario

### 🔄 Shifts

- **CRUD Operations**: Gestión de turnos
- **Filter by Store**: Turnos por tienda
- **Statistics**: Estadísticas de turnos
- **By Period**: Turnos por período

### 📊 Shift Planning

- **Preview**: Vista previa de proyección de turnos
- **Summary**: Resumen de proyección
- **Projected Days**: Días proyectados

### 🚫 Absences

- **Create**: Registrar ausencias
- **By Employee**: Ausencias por empleado
- **By Date Range**: Ausencias por rango de fechas
- **Total Hours**: Total de horas de ausencia

### 🎯 Novelties

- **Create**: Registrar novedades
- **By Employee**: Novedades por empleado
- **By Type**: Novedades por tipo
- **Summary**: Resumen de tipos de novedades

## 🔄 Flujo de Trabajo Recomendado

1. **Configurar servidor**: Asegúrate de que tu API esté ejecutándose en `http://localhost:8080`
2. **Login**: Ejecuta el endpoint de login para obtener el token
3. **Seleccionar contexto**: Si es necesario, selecciona franquicia/tienda
4. **Probar endpoints**: Ejecuta los endpoints en el orden lógico según tu caso de uso

## ⚠️ Notas Importantes

### Autenticación

- La mayoría de endpoints requieren autenticación JWT
- El token se incluye automáticamente en las headers después del login
- Algunos endpoints requieren roles específicos (admin)

### Middleware de Seguridad

- **JWT Middleware**: Validación de token en endpoints protegidos
- **Role Middleware**: Validación de roles (`admin` requerido para algunos endpoints)
- **Franchise Access**: Validación de acceso a franquicia específica

### Parámetros Comunes

- **Fechas**: Formato `YYYY-MM-DD` (ej: `2025-01-15`)
- **Horas**: Formato `HH:MM:SS` (ej: `08:00:00`)
- **IDs**: Números enteros positivos

## 🐛 Solución de Problemas

### Error 401 - Unauthorized

- Verificar que el token esté configurado correctamente
- Re-ejecutar el login si el token expiró

### Error 403 - Forbidden

- Verificar que el usuario tenga los roles necesarios
- Algunos endpoints requieren rol `admin`

### Error 400 - Bad Request

- Verificar el formato de los parámetros
- Revisar que todos los campos requeridos estén presentes

### Error 404 - Not Found

- Verificar que los IDs existan en la base de datos
- Comprobar la URL del endpoint

## 📝 Ejemplos de Datos

### Usuario/Empleado

```json
{
  "first_name": "Juan",
  "last_name": "Pérez",
  "document_type": "CC",
  "document_number": "12345678",
  "birthdate": "1990-01-01",
  "phone": "3001234567",
  "email": "juan.perez@example.com",
  "position": "Vendedor",
  "password": "password123",
  "salary": 1200000,
  "role_id": 2,
  "franchise_id": 1
}
```

### Turno

```json
{
  "employee_id": 1,
  "store_id": 1,
  "shift_date": "2025-01-15",
  "start_time": "08:00:00",
  "end_time": "17:00:00",
  "description": "Turno regular"
}
```

### Ausencia

```json
{
  "employee_id": 1,
  "absence_date": "2025-01-15",
  "start_time": "08:00:00",
  "end_time": "12:00:00",
  "reason": "Cita médica",
  "description": "Consulta médica general"
}
```

### Novedad

```json
{
  "employee_id": 1,
  "novelty_date": "2025-01-15",
  "novelty_type": "overtime",
  "start_time": "18:00:00",
  "end_time": "20:00:00",
  "description": "Horas extra por evento especial",
  "hours": 2
}
```

## 🔗 Enlaces Útiles

- [Documentación de Postman](https://learning.postman.com/docs/getting-started/introduction/)
- [JWT.io](https://jwt.io/) - Para decodificar tokens JWT
- [Postman Learning Center](https://learning.postman.com/) - Tutoriales y guías

---

¡Feliz testing! 🚀
