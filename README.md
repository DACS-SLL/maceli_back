# MACELI Backend API

Backend REST para MACELI, una marca de comida saludable en Arequipa. La API permite listar planes, registrar pedidos, recibir mensajes de contacto y administrar la informacion principal del negocio.

El proyecto esta pensado como un MVP claro. Usa una arquitectura simple por capas:

- `cmd/server`: punto de entrada de la aplicacion.
- `internal/config`: lectura de variables de entorno.
- `internal/database`: conexion, migraciones y seed inicial.
- `internal/models`: entidades de la base de datos.
- `internal/handlers`: controladores HTTP.
- `internal/routes`: definicion de rutas y CORS.
- `internal/middleware`: proteccion simple para rutas administrativas.

## Tecnologias usadas

- Golang
- Gin
- GORM
- PostgreSQL, compatible con NeonDB usando `DATABASE_URL`
- godotenv
- gin-contrib/cors

## Instalacion

Desde la carpeta `backend`:

```bash
go mod tidy
```

## Configuracion del entorno

Copia el archivo de ejemplo:

```bash
cp .env.example .env
```

Variables:

```env
PORT=8080
DATABASE_URL=postgres://usuario:password@localhost:5432/maceli_db?sslmode=disable
FRONTEND_URL=http://localhost:5173
ADMIN_KEY=maceli_admin_123
```
## Ejecutar el backend

Desde la carpeta `backend`:

```bash
go run ./cmd/server
```

La API quedara disponible en:

```text
http://localhost:8080/api
```

Al iniciar, GORM ejecuta `AutoMigrate` para crear las tablas:

- `planes`
- `pedidos`
- `contactos`

Tambien se insertan planes iniciales solo si la tabla de planes esta vacia.

## Header administrativo

Las rutas administrativas requieren el header:

```http
X-ADMIN-KEY: maceli_admin_123
```

El valor real debe coincidir con la variable `ADMIN_KEY` del archivo `.env`.

## Endpoints

### Health check

```http
GET /api/health
```

Respuesta:

```json
{
  "status": "ok",
  "message": "MACELI API funcionando"
}
```

### Planes publicos

```http
GET /api/planes
GET /api/planes/:id
```

Devuelven solo planes activos.

### Administracion de planes

Requieren `X-ADMIN-KEY`.

```http
GET /api/admin/planes
POST /api/admin/planes
PUT /api/admin/planes/:id
PATCH /api/admin/planes/:id/desactivar
```

Crear plan:

```json
{
  "nombre": "Plan semanal",
  "descripcion": "Plan saludable de 7 dias.",
  "precio": 91,
  "categoria": "Plan semanal",
  "imagen_url": "/uploads/imagen.png",
  "activo": true
}
```

Tambien puedes crear un plan enviando `multipart/form-data`. En ese caso, si mandas el campo `imagen`, el backend guarda la imagen y asigna `imagen_url` automaticamente.

Campos para `multipart/form-data`:

```text
nombre
descripcion
precio
categoria
activo
imagen
```

Actualizar plan:

```json
{
  "nombre": "Plan semanal premium",
  "precio": 99,
  "activo": true
}
```

Para actualizar un plan con imagen, usa tambien `multipart/form-data` en:

```http
PUT /api/admin/planes/:id
```

Envia el archivo en el campo:

```text
imagen
```

El plan quedara actualizado con la nueva URL de imagen automaticamente.

### Subida simple de imagenes

Requiere `X-ADMIN-KEY`.

```http
POST /api/admin/upload
```

Usa `multipart/form-data` con el campo:

```text
imagen
```

Respuesta:

```json
{
  "message": "Imagen subida correctamente",
  "imagen_url": "/uploads/1760000000000000000.png"
}
```

Este endpoint queda disponible si quieres subir una imagen por separado, pero para planes ya no es necesario hacerlo manualmente: `POST /api/admin/planes` y `PUT /api/admin/planes/:id` aceptan el campo `imagen` y guardan `imagen_url` automaticamente.

### Pedidos

```http
POST /api/pedidos
```

Request:

```json
{
  "nombre_cliente": "Lucia Ramos",
  "telefono": "987654321",
  "plan_id": 1,
  "mensaje": "Quiero informacion por WhatsApp",
  "direccion_zona": "Yanahuara"
}
```

Rutas administrativas:

```http
GET /api/admin/pedidos
PATCH /api/admin/pedidos/:id/estado
```

Actualizar estado:

```json
{
  "estado": "confirmado"
}
```

Estados permitidos:

- `pendiente`
- `contactado`
- `confirmado`
- `cancelado`

### Contacto

```http
POST /api/contacto
```

Request:

```json
{
  "nombre": "Mario",
  "telefono": "999888777",
  "correo": "mario@example.com",
  "mensaje": "Deseo conocer los planes mensuales"
}
```

Ruta administrativa:

```http
GET /api/admin/contacto
```

## Respuestas JSON

Exito:

```json
{
  "message": "Pedido registrado correctamente",
  "data": {}
}
```

Error:

```json
{
  "error": "El telefono es obligatorio"
}
```

## Probar con Postman

1. Ejecuta el servidor con `go run ./cmd/server`.
2. Crea una variable `base_url` con `http://localhost:8080/api`.
3. Prueba `GET {{base_url}}/health`.
4. Para rutas admin agrega el header `X-ADMIN-KEY` con el valor de tu `.env`.
5. Crea un plan con `POST {{base_url}}/admin/planes`.
6. Lista planes publicos con `GET {{base_url}}/planes`.
7. Registra un pedido con `POST {{base_url}}/pedidos`.
8. Revisa pedidos con `GET {{base_url}}/admin/pedidos`.

## Notas del MVP

- No incluye login ni JWT.
- No incluye pagos en linea.
- La proteccion administrativa usa un header simple para facilitar la presentacion del proyecto.
- CORS permite el frontend definido en `FRONTEND_URL`; si no existe, usa `http://localhost:5173`.
