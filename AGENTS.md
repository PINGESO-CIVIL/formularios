# AGENTS.md — formularios

Guía para agentes de código (y personas) que trabajan en este repositorio.
`apps/api/` y `apps/web/` tienen su propio `AGENTS.md` con reglas específicas;
cuando trabajes dentro de una de ellas, ese archivo manda en lo técnico, pero
**nunca contradice el DAS**.

## Fuentes de verdad (en orden de prioridad)

1. **DAS — Desarrollo de Microservicios** ([`docs/referencia/das.md`](docs/referencia/das.md)):
   stack, versiones exactas, estructura, contenedores, contrato, seguridad y pipeline.
   **Lo que el DAS marca como obligatorio no se discute** y lo verifica el pipeline.
2. **Hitos 1 y 2** ([`docs/referencia/hitos-1-y-2.md`](docs/referencia/hitos-1-y-2.md)):
   checklist y formato de entrega.
3. **Documento de componentes de la sección** (pendiente de incorporar): nombre del
   servicio, número N, dependencias y mocks.
4. Lineamientos iniciales ([`docs/referencia/lineamientos.pdf`](docs/referencia/lineamientos.pdf)):
   solo como contexto; donde chocan con el DAS, gana el DAS (ver `ADR-0004`).
5. Requerimientos funcionales: [`docs/requerimientos.md`](docs/requerimientos.md) (`RF-FOR-xx`, `RN-FOR-xx`).

## Qué es este proyecto

Servicio **formularios** (PT4, Motor de Formularios) de la Línea de Productos de
Software del curso PINGESO (USACH).

**Objetivo:** API y microcomponentes para crear formularios dinámicos,
versionarlos y recolectar las respuestas de los usuarios.

Debe ser **lo más general posible**: un motor reutilizable por cualquier
producto que componga la base común (contratos, checklists, fichas, encuestas,
consentimientos…), sin conocer el negocio de ninguno. Las diferencias entre
productos se expresan como **configuración** (esquema del formulario, widgets
registrados, tema visual), nunca como código específico.

Cada cliente despliega **su propia instancia** del servicio (DAS §8: el ambiente
productivo lo define cada grupo cliente). Por eso el servicio **no es
multi-tenant**: el JWT de identidad no trae tenant (ver `ADR-0004`).

### Qué SÍ hace

- Definir formularios: estructura dinámica en JSON Schema + `esquema_ui`.
- Versionarlos: borrador → publicada. Una versión publicada es inmutable.
- Recibir respuestas sobre una versión publicada, validarlas en el servidor y
  guardarlas: borrador → enviada. Una respuesta enviada es inmutable.
- Registrar en su **tabla de auditoría** quién, qué y cuándo de cada cambio (DAS §8).
- Entregar microcomponentes React (renderizador, listados, editor) moldeables
  visualmente por cada producto.

### Qué NO hace (dueño único por dato, DAS §2)

| No hacemos                             | Dueño                            | Cómo nos relacionamos                                                |
| -------------------------------------- | -------------------------------- | -------------------------------------------------------------------- |
| Login, usuarios, roles, emisión de JWT | **usuarios-roles** (PT6)         | Validamos su JWT con `AUTH_JWKS_URL`. Guardamos solo el `sub` (UUID) |
| Guardar archivos, fotos o firmas       | Bóveda Documental (PT7)          | Guardaremos solo el identificador que devuelva                       |
| Generar PDF o Excel                    | Motor de Reportes (PT1)          | Nos consume por API                                                  |
| Plantillas de **documentos**           | Gestor de Plantillas (PT3)       | Ninguna; nosotros solo manejamos **formularios**                     |
| Enviar correos                         | Gateway de Notificaciones (PT10) | No lo llamamos                                                       |

Si una tarea empieza a parecerse a algo de la columna "Dueño", **detente y
pregunta**. Nunca nos conectamos a la base de otro servicio ni escribimos sus datos.

## Dependencias

| Servicio       | Para qué                       | Variable                                                       | Contrato                                 | Mientras no tenga lógica                                                                  |
| -------------- | ------------------------------ | -------------------------------------------------------------- | ---------------------------------------- | ----------------------------------------------------------------------------------------- |
| usuarios-roles | Validar JWT (RS256) y permisos | `AUTH_JWKS_URL` (y `USUARIOS_ROLES_URL` si se necesitan datos) | `contracts/externos/usuarios-roles.yaml` | Ya emite tokens reales de prueba: `POST /api/dev/tokens` (ver su `docs/jwt-de-prueba.md`) |

Las demás dependencias se agregan cuando el documento de componentes las fije.
Reglas (DAS §2):

- REST síncrono con JSON contra el contrato publicado del otro servicio.
- Llamadas por `<SERVICIO>_URL`, reenviando el JWT y el `X-Request-Id`.
- Timeout explícito en todo cliente HTTP (5 s). Un servicio caído devuelve un error claro.
- El contrato del proveedor se copia a `contracts/externos/<servicio>.yaml` y el
  cliente se genera desde ahí (oapi-codegen en Go, openapi-typescript en el front).
- Mientras el proveedor no tenga lógica: `prism mock` o handlers MSW.
- No inventes comportamiento de otro servicio: marca `// TODO(integracion): ...`
  y anótalo en `docs/preguntas-abiertas.md`.

## Stack (DAS §3, versiones exactas, sin rangos)

| Capa                | Tecnología                                                                                                                                                      |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| API                 | Go 1.26 + **Huma v2.34.1** (HTTP + OpenAPI 3.1 generado desde los structs)                                                                                      |
| Base de datos       | **PostgreSQL 17.11**, `pgx/v5` v5.10.0, **goose v3.24.1** (migraciones), **sqlc v1.29.0** (queries)                                                             |
| JWT                 | `golang-jwt/jwt/v5` v5.3.1 + `MicahParks/keyfunc/v3` v3.9.1                                                                                                     |
| Pruebas             | `testing` + `testcontainers-go/modules/postgres` v0.43.0 (e2e con PostgreSQL real)                                                                              |
| Frontend            | **React 19.2.8**, TypeScript 6.0.3, Vite 8.1.5, Tailwind 4.3.3, shadcn (`@base-ui/react`), react-query, react-hook-form, openapi-fetch, MSW, Vitest, Playwright |
| Herramientas        | Node 22.22.0, pnpm 10.13.1, oxlint, Prettier, lefthook, gitleaks                                                                                                |
| Calidad y seguridad | SonarQube, Trivy, govulncheck, Hadolint, Dockle, oasdiff                                                                                                        |

Cambiar una pieza del stack o subir una versión es **un PR propio**, nunca dentro
de un PR de funcionalidad. Las versiones de este repo deben coincidir con el DAS;
toda desviación queda registrada en un ADR.

## Estructura del repositorio (DAS §3)

```
formularios/
├── apps/
│   ├── api/
│   │   ├── cmd/api/             # main: arranque, config, señales, subcomando openapi
│   │   ├── internal/api/        # handlers Huma + structs del contrato
│   │   ├── internal/service/    # reglas de negocio, sin HTTP ni SQL
│   │   ├── internal/store/      # acceso a datos (memoria en Hito 1, PostgreSQL en Hito 2)
│   │   ├── internal/repository/ # generado por sqlc, no editar
│   │   ├── db/migrations/       # goose
│   │   ├── db/queries/          # SQL para sqlc
│   │   ├── e2e/                 # e2e + testdata/*.golden.md
│   │   ├── Dockerfile
│   │   └── Dockerfile.migrate
│   └── web/                     # microcomponentes React + app de desarrollo con MSW
├── contracts/
│   ├── openapi.yaml             # GENERADO con pnpm codegen y commiteado
│   ├── form-schema.md           # formato del esquema de formulario
│   └── externos/                # contratos de los servicios de los que dependemos
├── docs/
│   ├── modos-de-falla.md        # obligatorio: se escribe antes de programar
│   ├── preguntas-abiertas.md
│   ├── adr/
│   └── referencia/              # DAS, hitos, lineamientos (material del curso)
├── docker-compose.yaml
├── .env.example
├── lefthook.yml
└── .github/workflows/ci.yml
```

## Reglas de arquitectura

1. **API-first con Huma.** El contrato nace de los structs Go: primero structs y
   handlers con datos simulados, después la lógica real. `pnpm codegen` genera
   `contracts/openapi.yaml` y los tipos TypeScript; ambos se commitean. Nunca se
   edita el `openapi.yaml` a mano.
2. **Handler → Service → Store.** El handler solo traduce HTTP; el service no
   conoce HTTP ni SQL; el store no tiene reglas de negocio. **Un handler que llama
   al store se rechaza.**
3. **Dueño único por dato y dependencias en un solo sentido**, sin ciclos.
4. **Cambios incompatibles** del contrato los bloquea oasdiff. Si son necesarios,
   se publica `/api/v2/...` y se avisa a los consumidores.
5. **Microcomponentes** en el frontend (ver `apps/web/AGENTS.md`).

## Estándares

### Contrato y API

- OpenAPI 3.1 generado por Huma. Documentación interactiva en `/api/docs` con
  `API_ENV=dev` (y `test`); nunca en `prod`.
- Rutas en **español y plural**, con prefijo `/api`: `/api/formularios`,
  `/api/formularios/{id}/versiones`, `/api/respuestas`.
- Campos JSON en español y `snake_case` (`creado_en`, `formulario_id`).
- Identificadores expuestos: **UUID**.
- **Toda entrada validada en el contrato** (tipos, `minLength`, `maxLength`,
  `maximum`, `pattern`, `enum`). Huma rechaza lo inválido con 422 antes del service.
- **Cada operación declara su permiso** en `x-permiso`: `publico`, `autenticado` o
  `<recurso>:<accion>`. Sin token válido → 401; sin permiso → 403.
  Permisos propuestos: `formularios:administrar`, `respuestas:operar`,
  `respuestas:administrar` (se cargan como `permisos_extra` en los tokens de prueba).
- Autorizar **por permiso, nunca por nombre de rol**.
- Errores en **RFC 9457** (`application/problem+json`), sin trazas, SQL ni rutas internas.
- `Idempotency-Key` en operaciones que cambian datos de otro servicio.

### Datos

- Migraciones **goose** en `db/migrations/`. Una migración ya desplegada no se edita:
  se crea otra.
- Queries en `db/queries/`; `internal/repository/` lo genera sqlc y no se edita.
- Fechas `timestamptz`, montos `bigint`, ids `uuid`.
- La API **nunca migra al arrancar**: corre primero el contenedor `migrate`.
- Inmutabilidad: una versión publicada y una respuesta enviada no se modifican ni
  se borran físicamente.
- Auditoría: toda operación que cambia datos de negocio escribe en la tabla
  `auditoria` (quién, qué, cuándo), en la misma transacción.

### Configuración y secretos (DAS §5)

- Todo por variables de entorno, validadas al arrancar; el proceso termina si
  falta una obligatoria. `.env.example` con valores `replace-with-*`.
- Variables: `API_PORT` (8080), `API_ENV` (`dev|test|prod`), `DATABASE_URL`,
  `AUTH_JWKS_URL`, `FRONTEND_URL`, `TZ` (`America/Santiago`), `<SERVICIO>_URL`.
- Ningún secreto en el repo, la imagen, un `ARG` ni variables `VITE_*`.
- Sesión en cookie `HttpOnly`/`Secure`/`SameSite=Lax`; **`localStorage` y
  `sessionStorage` prohibidos** en el frontend.

### Observabilidad (DAS §8)

- Logs `log/slog` JSON a stdout con `servicio`, `version`, `request_id`, `metodo`,
  `ruta`, `estado`, `duracion_ms`.
- **Nunca** se loguean contraseñas, tokens, cookies, ni el cuerpo de requests o
  el contenido de respuestas de formularios. Solo IDs.
- `X-Request-Id` se propaga en las llamadas a otros servicios.
- `/healthz` (vivo) y `/readyz` (200 solo si PostgreSQL contesta), sin autenticación.

### Calidad y seguridad (DAS §6 y §7)

- Pipeline `ci.yml` con jobs **`calidad`**, **`seguridad`** e **`imagen`**; los tres
  son checks requeridos. Un PR que no pasa el pipeline no se integra.
- Umbrales del Hito 2: SonarQube (código nuevo) 0 bugs, 0 vulnerabilidades, 100 %
  hotspots revisados, duplicación < 3 %, cobertura ≥ 60 % (medida desde los e2e);
  Trivy 0 HIGH/CRITICAL con corrección; Gitleaks 0 hallazgos.
- Si el gate falla se corrige la causa: no se baja el umbral ni se excluyen archivos.
- Excepciones de Trivy solo en `.trivyignore.yaml` con CVE, motivo y fecha, aprobadas.

### Pruebas (DAS §9)

- **Primero `docs/modos-de-falla.md`**: cada falla posible, qué la impide y cómo se verifica.
- E2E como mecanismo principal: binario real de la API contra PostgreSQL real
  (Testcontainers o `TEST_DATABASE_URL`), escenario de dificultad media o alta,
  reporte determinista comparado con un golden en `e2e/testdata/`.
- Antes de aceptar un e2e se rompe a propósito la protección y se confirma que falla.
- Lógica con fechas probada también con `TZ=UTC` y `TZ=Pacific/Auckland`.
- Matriz de acceso por permiso en `acceso_test.go`.

### Idioma y formato (DAS §9)

- **Identificadores y comentarios en español**, en Go y en TypeScript.
- Prettier sin punto y coma, ancho 90; gofmt en Go. Lo aplica lefthook.
- Dependencias JS en el `catalog` de `pnpm-workspace.yaml`, nunca sueltas.

## Hitos

| Hito                      | Fecha              | Se aprueba con                                                                                                                                     |
| ------------------------- | ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1. Implementación inicial | **4 de diciembre** | Contrato publicado y validado, endpoints con datos simulados, Compose levanta todo, `docs/modos-de-falla.md`, revisión con nota de la arquitectura |
| 2. Refinamiento y calidad | **4 de diciembre** | Lógica real, PostgreSQL, auditoría, autorización por permiso, e2e con golden, umbrales de SonarQube/Trivy/Gitleaks                                 |
| Ensamblaje LPS            | 21 de diciembre    | APIs y microcomponentes listos para que cada grupo cliente los componga                                                                            |
| Cierre                    | 11 de enero        | Validación con clientes                                                                                                                            |

Entrega de cada hito: tag `hito-N` en `main` con pipeline verde → release con
`openapi.yaml` (y SBOM en el Hito 2) → issue "Entrega Hito N" con la checklist.
Los tags quedan protegidos.

## Flujo de trabajo

1. **Plan → verificación → implementación.** Antes de cambios no triviales,
   escribe un plan y valídalo con el equipo.
2. **Nada sobre el repo sin OK.** Commits, push, amend, rebase, tags, releases o
   mover/borrar archivos versionados requieren aprobación explícita del equipo.
3. **Ramas:** `feat/<tema>`, `fix/<tema>`, `docs/<tema>`, `chore/<tema>`, una por issue.
   `main` protegida: merge solo por PR con una aprobación y checks en verde.
4. **Commits:** Conventional Commits, `tipo: mensaje corto y representativo`
   (`feat`, `fix`, `docs`, `chore`, `test`, `refactor`, `ci`).
   **Sin `Co-Authored-By` ni líneas de atribución.**
5. Un PR = una pieza verificable. Decisiones transversales → ADR en `docs/adr/`.
6. Antes de terminar: `pnpm verify`, `pnpm typecheck`, `pnpm test`, `pnpm test:api`
   y `pnpm codegen` sin diferencias.

## Definición de terminado

- Pipeline en verde (`calidad`, `seguridad`, `imagen`).
- `pnpm codegen` sin diferencias y oasdiff sin cambios incompatibles contra `main`.
- `docs/modos-de-falla.md` actualizado con las fallas nuevas.
- Pruebas del caso feliz y de al menos un rechazo por regla.
- Sin `TODO` nuevos salvo `TODO(integracion)` registrados en preguntas abiertas.
