# Research: Migración de Credenciales a GCP Secret Manager

**Feature**: 002-migrate-secrets-gcp
**Fecha**: 2026-05-03

---

## Decisión 1: Mecanismo de inyección en App Engine Standard

**Decisión**: Usar `secret_env_variables` en `app.yaml` (bloque nativo de App Engine).

**Fundamento**: App Engine Standard soporta `secret_env_variables` desde 2021. El runtime `go122` es compatible. En tiempo de deploy, App Engine resuelve cada referencia accediendo a Secret Manager y la inyecta como variable de entorno en la instancia — comportamiento idéntico al de `env_variables` pero con el valor almacenado en Secret Manager, no en el archivo de configuración. `pkg/config/config.go` no requiere cambios.

**Alternativas descartadas**:
- Script de deploy con `gcloud secrets versions access` + generación dinámica de `app.yaml`: más frágil, requiere lógica adicional en CI/CD, superficie de ataque mayor.
- Secret Manager SDK en tiempo de ejecución (`cloud.google.com/go/secretmanager`): agrega dependencia de código, requiere latencia de red al startup, cambia la interfaz del config.

**Sintaxis en `app.yaml`**:
```yaml
secret_env_variables:
  - name: DB_PASSWORD
    secret: loopi-db-password
    version: latest
  - name: JWT_SECRET
    secret: loopi-jwt-secret
    version: latest
```

**Referencia**: [App Engine Standard — Using secrets](https://cloud.google.com/appengine/docs/standard/reference/app-yaml#secret-env-variables)

---

## Decisión 2: Nomenclatura de los secretos en Secret Manager

**Decisión**: Prefijo `loopi-` + nombre descriptivo en kebab-case.

| Variable de entorno | Nombre del secreto en Secret Manager |
|---------------------|---------------------------------------|
| `DB_PASSWORD`       | `loopi-db-password`                   |
| `JWT_SECRET`        | `loopi-jwt-secret`                    |

**Fundamento**: El prefijo `loopi-` evita colisiones si el proyecto GCP hospeda otros servicios. El formato kebab-case es el estándar de Secret Manager.

---

## Decisión 3: Cuenta de servicio de App Engine

**Decisión**: La cuenta de servicio predeterminada de App Engine para el proyecto `quotes-api-100` es `quotes-api-100@appspot.gserviceaccount.com`.

**Fundamento**: En App Engine Standard, el runtime usa la cuenta de servicio predeterminada del proyecto (`PROJECT_ID@appspot.gserviceaccount.com`) a menos que se configure una cuenta personalizada. El proyecto ya usa esta cuenta para otros recursos (Cloud SQL).

**Rol requerido**: `roles/secretmanager.secretAccessor` a nivel de proyecto.

**Comando**:
```bash
gcloud projects add-iam-policy-binding quotes-api-100 \
  --member="serviceAccount:quotes-api-100@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

---

## Decisión 4: Rotación de credenciales comprometidas

**Decisión**: Proceso manual en dos pasos: (1) generar nuevos valores fuertes localmente, (2) subirlos a Secret Manager antes de actualizar `app.yaml`. Los valores anteriores quedan inválidos al rotarlos en la base de datos y en el sistema de autenticación.

**Pasos de rotación**:

### DB_PASSWORD
1. Conectarse a Cloud SQL y cambiar la contraseña del usuario `loopi-user`:
   ```sql
   ALTER USER 'loopi-user'@'%' IDENTIFIED BY 'NUEVA_CONTRASEÑA_FUERTE';
   FLUSH PRIVILEGES;
   ```
2. Subir el nuevo valor a Secret Manager:
   ```bash
   echo -n 'NUEVA_CONTRASEÑA_FUERTE' | gcloud secrets versions add loopi-db-password \
     --data-file=- --project=quotes-api-100
   ```

### JWT_SECRET
1. Generar un nuevo secret aleatorio de 256 bits:
   ```bash
   openssl rand -base64 32
   ```
2. Subir a Secret Manager:
   ```bash
   echo -n 'NUEVO_JWT_SECRET' | gcloud secrets versions add loopi-jwt-secret \
     --data-file=- --project=quotes-api-100
   ```
3. **Efecto**: todos los tokens JWT activos quedarán inválidos (los usuarios deberán hacer login nuevamente). Es un efecto aceptable y esperado.

---

## Decisión 5: Variables no sensibles en `app.yaml`

**Decisión**: Las variables no sensibles (`SERVER_PORT`, `APP_ENV`, `DB_INSTANCE_CONNECTION`, `DB_USER`, `JWT_EXPIRATION_HOURS`, `TZ`, `LOG_LEVEL`, `LOG_FORMAT`, `SERVICE_NAME`, `SERVICE_VERSION`, `SERVICE_ENV`) permanecen en `env_variables` — no son secretos y no requieren Secret Manager.

**Fundamento**: Secret Manager es para valores que no deben aparecer en logs, repositorios ni consolas. `DB_USER`, `DB_INSTANCE_CONNECTION`, etc. son configuración de conexión no sensible.

---

## Decisión 6: Desarrollo local

**Decisión**: `.env` local (gitignoreado) sigue siendo el mecanismo de desarrollo. No se integra con Secret Manager en este alcance.

**Fundamento**: El acceso a Secret Manager desde local requiere autenticación ADC (`gcloud auth application-default login`) y latencia de red — innecesario cuando `.env` cumple la misma función de forma más simple. El `.env.example` debe documentar cómo obtener los valores para el archivo local.
