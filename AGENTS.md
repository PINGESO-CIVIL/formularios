# AGENTS.md — ms-formularios

Guía para agentes de código (y personas) que trabajan en este repositorio.
Las carpetas `backend/` y `frontend/` tienen su propio `AGENTS.md` con reglas
específicas; cuando trabajes dentro de una de ellas, ese archivo manda sobre este
en lo técnico, pero **nunca contradice los lineamientos del curso**.

> **Fuente de verdad técnica:** [`docs/referencia/lineamientos.pdf`](docs/referencia/lineamientos.pdf)
> ("Lineamientos Técnicos y Plan Incremental"). Si algo de este repo choca con
> ese documento, gana el documento y se corrige el repo.
> **Fuente de verdad funcional:** [`docs/requerimientos.md`](docs/requerimientos.md)
> (IDs `RF-FOR-xx`, `RNF-FOR-xx`, `RN-FOR-xx`).

## Qué es este proyecto

`ms-formularios` es el **Paquete de Trabajo 4 (PT4): Motor de Formularios** de
la Línea de Productos de Software (LPS) del curso Proyecto de Ingeniería en
Software (PINGESO, USACH).

**Objetivo (lineamientos):** API para crear formularios dinámicos y recolectar
las respuestas de los usuarios.

Debe ser **lo más general posible**: un motor reutilizable por cualquier
producto derivado de la LPS (contratos, checklists, fichas, encuestas,
consentimientos, autoevaluaciones…), sin conocer el negocio de ninguno. Cada
cliente/producto es un *tenant*; sus diferencias se expresan como
configuración (features por tenant, tipos de campo registrados, esquema del
formulario), nunca como código específico.

### Qué SÍ hace ms-formularios

- Definir formularios (estructura dinámica en JSON Schema + UI schema).
- Versionarlos: borrador → publicado → archivado. Una versión publicada es inmutable.
- Recibir respuestas sobre una versión publicada, validarlas en el servidor y
  guardarlas (borrador → enviada). Una respuesta enviada es inmutable.
- Consultar y listar formularios, versiones y respuestas, filtradas por tenant.
- Emitir un evento/log estructurado por cada acción relevante (para PT8).

### Qué NO hace (es de otro paquete de trabajo)

| No hacemos | Lo hace | Cómo nos relacionamos |
| --- | --- | --- |
| Login, usuarios, roles, emisión de JWT | **PT6 Identity Provider** | Solo **validamos** el JWT y leemos sus claims |
| Guardar archivos, fotos, firmas digitales de documentos | **PT7 Bóveda Documental** | Guardamos solo el `file_id` que devuelve PT7 |
| Bitácora / historial / dashboard de auditoría | **PT8 Central Auditoría** | Le enviamos logs estructurados de eventos |
| Generar PDF o Excel de una respuesta | **PT1 Motor de Reportes** | PT1 consume nuestros datos por API |
| Plantillas de **documentos** (estructuras de reportes/documentos) | **PT3 Gestor de Plantillas** | Ninguna en fase 1. Nosotros solo manejamos definiciones de **formularios** |
| Enviar correos | **PT10 Gateway de Notificaciones** | No lo llamamos en fase 1 |
| Reservas, OCR, ETL masivo | PT9, PT5, PT2 | Posibles consumidores de nuestra API; no construimos nada para ellos |

Si una tarea empieza a parecerse a algo de la columna "Lo hace", **detente y
pregunta**: no invadas otro MS.

## Ecosistema LPS (resumen de lineamientos)

| PT | Grupo | Objetivo |
| --- | --- | --- |
| PT1 | Motor de Reportes | Recibe datos crudos + ID de plantilla y devuelve PDF/Excel |
| PT2 | Pipeline de Información | ETL: recibir, limpiar, validar e insertar datos masivos |
| PT3 | Gestor de Plantillas | CRUD para crear, versionar y almacenar estructuras de documentos |
| **PT4** | **Motor de Formularios (nosotros)** | **API para crear formularios dinámicos y recolectar respuestas** |
| PT5 | Extractor OCR | Imagen digitalizada → texto OCR → JSON |
| PT6 | Identity Provider | Autenticación, login, roles y emisión de JWT |
| PT7 | Bóveda Documental | Recibe archivos, los almacena seguros, firma y guarda metadata |
| PT8 | Central Auditoría | Recibe logs estructurados de eventos y muestra dashboard |
| PT9 | Motor de Agendamiento | Reservas, colisiones, bloqueo de fechas |
| PT10 | Gateway de Notificaciones | Recibe (asunto, cuerpo) y envía email |

## Dependencias de ms-formularios

Dependencia oficial según lineamientos: **PT6**. Las demás son dependencias que
el equipo identificó por la naturaleza de un MS de formularios.

| Dependencia | Tipo | Puerto (backend) | Mock en fase 1 |
| --- | --- | --- | --- |
| **PT6 Identity Provider** (oficial) | Entrante: valida cada petición | `ports.Authenticator` | Token JWT quemado / simulación de ID-token válido de usuario emisor |
| PT8 Central Auditoría | Saliente: un log por evento | `ports.EventPublisher` | Imprimir el evento en consola (JSON) |
| PT7 Bóveda Documental | Saliente: campos archivo/firma | `ports.FileStore` | `file_id` falso; se asume que el archivo existe y está autenticado |
| PT3 Gestor de Plantillas | Límite a coordinar | — | — (ver `docs/preguntas-abiertas.md`) |
| PT1, PT2, PT5, PT9 | Consumidores potenciales | — | — |

Reglas:

- Toda dependencia externa es una **interfaz (puerto)** con un **mock local**.
  Integrar con el MS real = agregar un adaptador nuevo, sin tocar el dominio.
- **No inventes** el comportamiento de otro MS (formato real del token, nombres
  de roles, URL, formato de log de PT8). Déjalo en el adaptador, marca
  `// TODO(integracion): ...` y anótalo en `docs/preguntas-abiertas.md`.
- Entidades de negocio de los productos (cliente, vehículo, paciente…) se
  referencian como `subject_type` + `subject_id` opacos; nunca se validan
  contra otro servicio.

## Stack oficial (obligatorio por lineamientos)

| Capa | Tecnología | Versión |
| --- | --- | --- |
| Contrato API | OpenAPI (YAML) + Swagger UI | **3.1.0** |
| Backend | Go | **1.23** |
| Frontend | Vue.js, Composition API | **3.5** |
| Base de datos | **PostgreSQL** (`JSONB` para esquemas y respuestas dinámicas) | **16** |
| Análisis estático | **SonarQube** en el pipeline (bugs, code smells, cobertura) | — |
| Vulnerabilidades | **Trivy** (contenedores y dependencias Go/Vue) | — |

> Decisión de base de datos documentada en [`docs/adr/ADR-0003-persistencia-postgresql-jsonb.md`](docs/adr/ADR-0003-persistencia-postgresql-jsonb.md):
> se utiliza PostgreSQL 16 (`pgx/v5`, `sqlc`, `golang-migrate`) con columnas `JSONB` para estructuras dinámicas.

## Reglas de arquitectura (lineamientos)

1. **API-First.** El contrato `contracts/openapi.yaml` se publica **antes** de
   programar. Cualquier cambio de API se hace primero en el contrato y luego en
   backend y frontend. Un PR que cambia un endpoint sin tocar el contrato se rechaza.
2. **Clean Architecture: Handlers → Services → Repositories.**
   - *Handlers*: HTTP; parsean, validan forma, llaman al service, traducen respuesta.
   - *Services*: casos de uso y reglas de negocio.
   - *Repositories*: acceso a PostgreSQL.
   - El dominio no conoce HTTP ni PostgreSQL. (Ver `backend/AGENTS.md` para la
     correspondencia con las carpetas.)
3. **Frontends modulares.** Vistas y componentes aislados, fáciles de ensamblar
   en el producto final de la LPS.
4. **Mocks para trabajo paralelo.** Los grupos que dependen de nosotros deben
   poder simular nuestras respuestas leyendo solo el contrato; nosotros hacemos
   lo mismo con PT6/PT7/PT8.

## Estándares

### Contrato (OpenAPI)

- OpenAPI **3.1.0**, archivo `contracts/openapi.yaml`, documentado con Swagger UI.
- El estilo sigue el ejemplo entregado por el equipo docente
  ([`docs/referencia/swagger-ejemplo.yaml`](docs/referencia/swagger-ejemplo.yaml)).
  Ese archivo es **solo un ejemplo de estilo** (gestor de tareas), no nuestro contrato.
- Rutas versionadas con prefijo `/api/v1`, recursos en **español y plural**
  (`/api/v1/formularios`, `/api/v1/formularios/{id}/versiones`, `/api/v1/respuestas`).
- Campos JSON en **español y `snake_case`** (`fecha_creacion`, `tenant_id`).
- Esquemas separados de entrada y salida: `XxxRequest` / `XxxResponse`
  (la respuesta incluye lo enviado + campos autogenerados: `id`, `fecha_creacion`…).
- IDs: **UUID** (`type: string`, `format: uuid`).
- Fechas en ISO 8601 / RFC 3339, UTC (`format: date-time`).
- Errores con el esquema estándar del ejemplo, reutilizado vía `components/responses`
  (`BadRequest`, `NotFound`, `InternalServerError`, más `Unauthorized`, `Forbidden`, `Conflict`):

  ```json
  { "codigo": 400, "mensaje": "Datos de entrada inválidos.", "detalles": "El campo 'titulo' es obligatorio." }
  ```

  Para errores de validación de una respuesta de formulario se permite el campo
  opcional `errores: [{ "campo": "...", "mensaje": "..." }]`. Mensajes en español.
- Todo endpoint (salvo salud) exige `Authorization: Bearer <JWT de PT6>`.
- Todo endpoint tiene `tags`, `summary`, ejemplos y todas sus respuestas de error.

### Seguridad y datos

- **Multi-tenant siempre.** Todo documento de negocio tiene `tenant_id` y toda
  consulta filtra por él. El `tenant_id` sale del token, **nunca** del cuerpo.
- **Inmutabilidad.** Una versión publicada y una respuesta enviada no se
  modifican ni se borran físicamente (RN-FOR-01, RN-FOR-04).
- **Sin bitácora propia.** El historial lo arma PT8 a partir de nuestros eventos.
- **Datos sensibles.** No registres en logs el contenido de respuestas, RUT,
  firmas ni tokens. Solo IDs.
- **Sin lógica de un producto en el núcleo.** Nada de `if tenant == "x"`.
- Secretos solo por variables de entorno; nunca en el repo.

### Calidad

- **SonarQube** corre en el pipeline; el PR no se mezcla si no pasa el quality gate.
- **Trivy** escanea imagen Docker y dependencias (Go y pnpm); sin vulnerabilidades
  `HIGH`/`CRITICAL` sin justificar.
- Pruebas: caso feliz + al menos un caso de error por regla. Cobertura dentro del
  umbral que defina el curso en SonarQube.
- Lint y formato: `gofmt` + `golangci-lint` en backend; ESLint + Prettier en frontend.

### Idioma

- Código Go/TS, identificadores, nombres de tablas y columnas SQL: **inglés**.
- Contrato público (rutas y campos JSON), textos de UI, mensajes de error y
  documentación: **español**.
- Eventos en formato LPS en español: `formulario.creado`, `version.publicada`,
  `respuesta.enviada`.

## Hitos (lineamientos)

| Hito | Fecha | Entregables |
| --- | --- | --- |
| Hito 0: Entorno y seguridad | (actual) | Repo oficial activo, Trivy integrado, SonarQube operativo |
| **Hito 1, Fase 1: Implementación inicial** | **13 de noviembre** | **Contrato publicado y Swagger validado; endpoints respondiendo con datos simulados (mocks)**; revisión de arquitectura con nota |
| Hito 2, Fase 2: Refinamiento y calidad | 4 de diciembre | Lógica real (cero mocks), persistencia en BD funcionando, umbrales de Trivy y SonarQube superados |
| Fase 3: Producción y cierre | 21 dic / 4 ene / 11 ene | Derivación LPS (ensamblaje) · despliegue cloud por cliente · validación con clientes |

**Foco actual: Hito 1.** Prioriza el contrato completo y validado, y endpoints
que respondan con mocks coherentes con él. La lógica real y la persistencia
son del Hito 2.

## Mapa del repositorio

```
.
├── AGENTS.md                 # este archivo
├── backend/                  # servicio Go (ver backend/AGENTS.md)
├── frontend/                 # monorepo Vue (ver frontend/AGENTS.md)
├── contracts/
│   ├── openapi.yaml          # contrato REST: fuente de verdad para back y front
│   ├── form-schema.md        # formato del esquema de formulario (JSON Schema + UI schema)
│   └── events/               # JSON Schema de cada evento y del sobre estándar
├── docs/
│   ├── referencia/
│   │   ├── lineamientos.pdf        # lineamientos técnicos del curso
│   │   └── swagger-ejemplo.yaml    # ejemplo de estilo OpenAPI entregado por el curso
│   ├── requerimientos.md
│   ├── preguntas-abiertas.md
│   └── adr/                  # decisiones de arquitectura (ADR-0001-....md)
├── docker-compose.yml        # infraestructura local
└── .github/workflows/        # CI (lint, pruebas, SonarQube, Trivy)
```

## Flujo de trabajo

1. **Plan → verificación → implementación.** Antes de cambios no triviales,
   escribe un plan y valídalo con el equipo. Recién entonces implementa.
2. **Nada sobre el repo sin OK.** Commits, push, amend, rebase o mover/borrar
   archivos versionados requieren aprobación explícita de una persona del equipo.
3. **Ramas:** `feat/RF-FOR-10-iniciar-respuesta`, `fix/...`, `docs/...`, `chore/...`.
4. **Commits:** `tipo: mensaje corto y representativo`, con el ID si aplica.
   Tipos: `feat`, `fix`, `docs`, `chore`, `test`, `refactor`.
   Ejemplo: `feat: iniciar respuesta sobre versión publicada (RF-FOR-10)`.
   **Sin `Co-Authored-By` ni líneas de atribución.**
5. **Pequeño y verificable.** Un PR = un requerimiento o una pieza de
   infraestructura, con pruebas.
6. Decisiones que afecten a más de un módulo → ADR corto en `docs/adr/`.
7. Antes de dar algo por terminado, corre las verificaciones de la carpeta
   (`make check` en backend, `pnpm check` en frontend).

## Definición de terminado

- Compila, lint sin errores, pruebas en verde, SonarQube y Trivy sin bloqueos.
- Contrato OpenAPI actualizado y válido si cambió la API.
- Pruebas que cubren el caso feliz y al menos un caso de error por regla.
- Sin `TODO` nuevos salvo `TODO(integracion)` registrados en preguntas abiertas.
- Cambios de esquema o índices en PostgreSQL con su migración `.up.sql` y `.down.sql` en `backend/db/migrations/`.
