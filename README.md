# MACELI Backend

API Go + Gin + PostgreSQL para el sitio de MACELI. Incluye contenido público, carta semanal, carga de imágenes con Cloudinary y administración con cuentas individuales.

La documentación vigente de configuración, seguridad, migraciones y despliegue está en [OPERACION.md](OPERACION.md).

## Desarrollo

1. Instalar la versión de Go indicada en `go.mod` o posterior compatible.
2. Copiar `.env.example` a `.env` y configurar una base PostgreSQL de desarrollo.
3. Definir `ADMIN_EMAIL` y `ADMIN_PASSWORD` para crear la primera cuenta.
4. Ejecutar `go mod download` y `go run ./cmd/server`.

Sin Cloudinary se usa almacenamiento local únicamente en desarrollo. En producción es obligatorio configurar Cloudinary.

## Validación

```sh
go vet ./...
go test -race ./...
go build ./cmd/server
```

La prueba de integración requiere `TEST_DATABASE_URL` con una base PostgreSQL desechable. Sin esa variable se omite; el workflow de GitHub incluye un servicio PostgreSQL dedicado.

La autenticación anterior por `X-ADMIN-KEY` fue retirada. No existen credenciales predeterminadas. Consulta los endpoints y el proceso de acceso en la guía de operación.
