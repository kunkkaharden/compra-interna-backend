# compra-interna-backend

Backend Go pa `compra-interna-frontend`: auth (usuario/contraseña/JWT),
catálogo de productos y lista mensual. Ver diseño completo en
`../compra-interna-frontend/docs/superpowers/specs/2026-07-01-compra-interna-backend-design.md`.

## Setup

```bash
cp .env.example .env   # editar JWT_SECRET
export $(cat .env | xargs)
go run ./cmd/api
```

## Crear primer usuario

No hay endpoint de registro público. Usar el CLI:

```bash
go run ./cmd/seed -usuario=admin -contrasenna=tu-clave
```

## Tests

```bash
go test ./...
```

## Endpoints

Ver spec de diseño para el listado completo. Resumen:

- `POST /api/auth/login`, `GET /api/auth/me`
- `GET|POST /api/users`, `GET|PUT|DELETE /api/users/:id`
- `GET|POST /api/products`, `PUT|DELETE /api/products/:id`
- `GET|POST /api/monthly-lists`, `GET|PUT|DELETE /api/monthly-lists/:id`

Todo protegido con `Authorization: Bearer <token>` excepto login y `/health`.
