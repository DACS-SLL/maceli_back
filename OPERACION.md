# MACELI: puesta en marcha del backend

## Configuración

Se conserva Go + Gin + PostgreSQL y Cloudinary. Revisar `go.mod` para la versión mínima de Go. `render.yaml` fija el runtime de Go utilizado para esta actualización.

- `DATABASE_URL`: conexión PostgreSQL existente con TLS en producción.
- `FRONTEND_URL`: origen HTTPS exacto de Vercel. Se permiten varios separados por comas; agregar localhost solo en desarrollo.
- `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`: credenciales privadas, solo en Render.
- `CLOUDINARY_UPLOAD_PATH`: carpeta de imágenes, sugerido `maceli/site`.
- `ADMIN_EMAIL`, `ADMIN_PASSWORD`: crean únicamente la primera cuenta. Contraseña única de 12 a 72 bytes UTF-8. Eliminarlas del entorno después del primer arranque correcto. No existe contraseña predeterminada ni registro público. Si ya hay cuentas, estas variables no modifican sus contraseñas.
- `APP_ENV=production`: exige Cloudinary para evitar pérdidas por almacenamiento efímero. Render también se detecta con `RENDER=true`.
- `TRUSTED_PROXIES`: rangos/IP de proxies verificados, separados por comas. Vacío no confía en cabeceras reenviadas. Nunca usar `0.0.0.0/0` ni `::/0` para resolver un límite de tráfico.

La autenticación anterior `X-ADMIN-KEY` fue retirada. Integraciones administrativas antiguas deberán usar una sesión. Las tablas existentes de planes, pedidos y contactos se conservan. Las migraciones agregan cuentas, sesiones, contenidos y medios. No se insertan nuevos planes/precios de demostración automáticamente.

## Orden de despliegue

1. Respaldar PostgreSQL y confirmar que existe una conexión de prueba independiente.
2. Ejecutar `go vet ./...`, `go test -race ./...` y `go build ./cmd/server` en un entorno compatible.
3. La prueba de integración usa exclusivamente `TEST_DATABASE_URL`, crea un esquema temporal propio y lo elimina al terminar. Nunca asignarle la base de producción. El workflow incluye PostgreSQL desechable.
4. Configurar las variables anteriores en Render. La primera contraseña no se guarda en Git ni se expone en logs.
5. Desplegar backend. Verificar `/api/health`, login y creación del borrador.
6. Configurar el origen de Render en Vercel y desplegar frontend.
7. Probar con una imagen real: subir, guardar, verificar que aún no aparece públicamente, previsualizar y publicar. Confirmar carta, fechas, zoom y WhatsApp en móvil. Probar logout y una solicitud sin sesión.

## API nueva

- `GET /api/site`: contenido publicado y estado de carta (`empty`, `upcoming`, `current`, `expired`). Solo devuelve páginas de la carta vigente.
- `POST /api/auth/login`: `{email,password}` devuelve token aleatorio y vencimiento. Usar `Authorization: Bearer TOKEN` en rutas administrativas.
- `GET /api/admin/site`: borrador, versión y última publicación.
- `PUT /api/admin/site`: `{content,version}`. Rechaza conflictos con 409 y URLs no registradas por la carga autenticada.
- `POST /api/admin/site/publish`: `{version}` publica el borrador de forma atómica.
- `POST /api/admin/upload`: multipart `imagen`; devuelve `{url,id}`.
- `POST /api/admin/logout`: revoca la sesión.
- `GET/POST /api/admin/users`, `DELETE /api/admin/users/:id`: listar, crear y desactivar cuentas. No permite desactivar la propia cuenta ni dejar cero administradores activos.
- `PUT /api/admin/password`: `{current,password}` cambia la propia contraseña y revoca las sesiones.

## Seguridad y límites

- Contraseñas bcrypt coste 12; tokens criptográficamente aleatorios de 256 bits, almacenados en la base solo como SHA-256. Sesiones de ocho horas, una sesión vigente por cuenta después de cada login.
- Límite global por proceso de 1800 solicitudes/minuto y 180 por IP. Login: 5 por IP y 30 globales/minuto. Cargas: 12 por IP y 30 globales/minuto. Pedidos/contactos: 5 por IP y 60 globales/minuto por ruta. Respuesta 429 con Retry-After.
- Memoria del limitador acotada a 10 000 entradas. En varias instancias, sustituir por un almacén compartido o límites equivalentes del perímetro. Con proxies no configurados, clientes pueden compartir el límite de IP: confirmar la topología antes de producción.
- Solicitudes JSON de hasta 128 KiB, multipart de hasta 9 MiB, cada imagen de hasta 8 MiB y 40 megapíxeles. Se inspecciona formato real y concordancia con extensión, se rechazan SVG y archivos no válidos. Subida a Cloudinary con timeout.
- Límites de lectura/escritura HTTP, contexto de solicitud de 35 segundos y pool PostgreSQL acotado. Cabeceras nosniff, DENY y no-store para administración; CORS de orígenes explícitos.
- La mitigación DDoS de Render protege su perímetro; los límites de aplicación no garantizan inmunidad. Revisar métricas, respuestas 429, consumo de Cloudinary y alarmas de infraestructura. No efectuar pruebas DDoS contra producción.

Fuentes: https://render.com/docs/ddos-protection y https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html

## Operación y recuperación

Guardar borrador y publicar son operaciones diferentes. La publicación anterior sigue intacta ante un error de validación/carga/guardado. Las fechas se calculan con horario de Lima. Los archivos retirados se conservan en Cloudinary; no hay eliminación física automática, para no romper referencias publicadas. Revisar periódicamente medios no usados y el consumo de almacenamiento.

Mantener al menos dos cuentas individuales de responsables. Si se pierde todo acceso, recuperar por un procedimiento de operador de base de datos: establecer un hash bcrypt válido y revocar todas las sesiones de la cuenta afectada. El bootstrap no actúa como puerta trasera ni restablecimiento automático. La entrega no incluye envío de contraseñas ni recuperación por correo.
