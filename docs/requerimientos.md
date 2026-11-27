# Requerimientos — formularios (PT4)

> Fuente de verdad funcional del servicio **formularios** (PT4 — Motor de
> Formularios) de la Línea de Productos de Software (LPS) PINGESO (USACH).
> Alineado con [`AGENTS.md`](../AGENTS.md), el
> [DAS](referencia/das.md), [Hitos 1 y 2](referencia/hitos-1-y-2.md) y
> [`ADR-0004`](adr/ADR-0004-adopcion-das.md).
>
> **Estado:** Propuesta para validación del equipo en PR (Hito 1). Los IDs
> `RF-FOR-xx`, `RNF-FOR-xx` y `RN-FOR-xx` se usan en commits, PRs, pruebas y en
> [`docs/modos-de-falla.md`](modos-de-falla.md). Se preservan los IDs ya
> referenciados en el repositorio (`RF-FOR-01`, `RF-FOR-04`, `RN-FOR-01`,
> `RN-FOR-04`).

Enlace al documento compartido del equipo: _pendiente de incorporar_

## Alcance y dueño único por dato (DAS §2)

**Objetivo:** proveer una API REST (Go + Huma) y microcomponentes React
moldeables (`@pingeso-civil/formularios-ui`) para definir formularios dinámicos,
versionarlos y recolectar respuestas validadas, sin acoplarse al negocio de
ningún producto cliente. Cada cliente despliega su propia instancia del
servicio (sin multi-tenancy; ver [`ADR-0004`](adr/ADR-0004-adopcion-das.md)).

| Responsabilidad fuera de alcance        | Servicio dueño                   | Relación con `formularios`                                       |
| --------------------------------------- | -------------------------------- | ---------------------------------------------------------------- |
| Login, usuarios, roles y emisión de JWT | **usuarios-roles** (PT6)         | Validamos su JWT RS256 vía `AUTH_JWKS_URL`; guardamos solo `sub` |
| Almacenamiento de archivos, fotos/firma | Bóveda Documental (PT7)          | Guardaremos solo el identificador opaco devuelto al integrarse   |
| Generación de reportes PDF o Excel      | Motor de Reportes (PT1)          | Consume respuestas desde nuestra API                             |
| Plantillas de documentos                | Gestor de Plantillas (PT3)       | Ninguna; este servicio gestiona solo formularios dinámicos       |
| Envío de correos y notificaciones       | Gateway de Notificaciones (PT10) | Ninguna; este servicio no lo invoca                              |

## Permisos declarados (`x-permiso`)

Cada endpoint declara su permiso en `x-permiso` (DAS §6). La autorización se
evalúa por permiso del claim `permisos` del JWT, nunca por nombre de rol.

| Valor de `x-permiso`      | Uso                                                                          |
| ------------------------- | ---------------------------------------------------------------------------- |
| `publico`                 | Endpoints operativos de salud (`/healthz`, `/readyz`), sin autenticación     |
| `autenticado`             | Lectura de formularios activos y versiones publicadas para usuarios con JWT  |
| `formularios:administrar` | Crear/editar/archivar formularios, gestionar borradores y publicar versiones |
| `respuestas:operar`       | Iniciar, guardar borrador, enviar y consultar las propias respuestas (`sub`) |
| `respuestas:administrar`  | Consultar y filtrar respuestas de todos los usuarios                         |

> Nota: los nombres de permisos propuestos requieren registro en el catálogo de
> `usuarios-roles` (PT6); ver [`docs/preguntas-abiertas.md`](preguntas-abiertas.md)
> (#3 y #15).

## Requerimientos funcionales (RF-FOR)

| ID        | Nombre                                     | Descripción                                                                                                                                                                   | Ruta / interfaz esperada                                                                   | Permiso (`x-permiso`)                          | Hito (mock → real) |
| --------- | ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ | ---------------------------------------------- | ------------------ |
| RF-FOR-01 | Crear formulario                           | Crear la definición maestra de un formulario dinámico con `nombre`, `descripcion`, `categoria` y `etiquetas`, en estado inicial `activo`, asociando `creado_por` (`sub`).     | `POST /api/formularios`                                                                    | `formularios:administrar`                      | H1 mock / H2 real  |
| RF-FOR-02 | Listar y consultar formularios             | Listar formularios con filtros básicos (`estado`, `categoria`, búsqueda/paginación) y obtener el detalle de un formulario por UUID.                                           | `GET /api/formularios`, `GET /api/formularios/{id}`                                        | `autenticado`                                  | H1 mock / H2 real  |
| RF-FOR-03 | Actualizar metadatos y archivar formulario | Modificar metadatos (`nombre`, `descripcion`, `categoria`, `etiquetas`) de un formulario y archivarlo lógicamente (`estado = archivado`) sin borrado físico.                  | `PATCH /api/formularios/{id}`, `POST /api/formularios/{id}/archivar`                       | `formularios:administrar`                      | H1 mock / H2 real  |
| RF-FOR-04 | Crear y editar borrador de versión         | Crear una nueva versión en estado `borrador` para un formulario y editar su `esquema` (JSON Schema 2020-12), `esquema_ui` y `notas` mientras permanezca en borrador.          | `POST /api/formularios/{id}/versiones`, `PUT /api/formularios/{id}/versiones/{version_id}` | `formularios:administrar`                      | H1 mock / H2 real  |
| RF-FOR-05 | Publicar versión de formulario             | Validar `esquema` y `esquema_ui`, asignar un número correlativo único, registrar `publicada_por` y `publicada_en` y pasar la versión a estado inmutable `publicada`.          | `POST /api/formularios/{id}/versiones/{version_id}/publicar`                               | `formularios:administrar`                      | H1 mock / H2 real  |
| RF-FOR-06 | Listar y consultar versiones               | Listar versiones de un formulario y consultar una versión específica (incluyendo la versión publicada vigente) por UUID para renderizado o inspección.                        | `GET /api/formularios/{id}/versiones`, `GET /api/formularios/{id}/versiones/{version_id}`  | `autenticado`                                  | H1 mock / H2 real  |
| RF-FOR-07 | Iniciar respuesta en borrador              | Crear una respuesta en estado `borrador` asociada a una versión `publicada` de un formulario `activo`, admitiendo referencia externa opaca (`sujeto_tipo`, `sujeto_id`).      | `POST /api/respuestas`                                                                     | `respuestas:operar`                            | H1 mock / H2 real  |
| RF-FOR-08 | Guardar avance de respuesta (borrador)     | Actualizar los `datos` JSON de una respuesta mientras permanezca en estado `borrador`, actualizando `actualizado_en`.                                                         | `PUT /api/respuestas/{id}`                                                                 | `respuestas:operar`                            | H1 mock / H2 real  |
| RF-FOR-09 | Enviar respuesta definitiva                | Validar `datos` en el servidor contra el JSON Schema 2020-12 de la versión (y formatos custom como `rut`), calcular `hash_contenido`, fijar `enviada_en` y pasar a `enviada`. | `POST /api/respuestas/{id}/enviar`                                                         | `respuestas:operar`                            | H1 mock / H2 real  |
| RF-FOR-10 | Listar y consultar respuestas              | Consultar una respuesta por UUID y listar respuestas con filtros (`formulario_id`, `version_id`, `estado`, `sujeto_tipo`, `sujeto_id`) y paginación; consumible por PT1.      | `GET /api/respuestas`, `GET /api/respuestas/{id}`                                          | `respuestas:operar` / `respuestas:administrar` | H1 mock / H2 real  |
| RF-FOR-11 | Registrar auditoría propia                 | Registrar `actor_id`, `accion`, `recurso`, `recurso_id`, `request_id` y `ocurrido_en` en la tabla `auditoria` dentro de la misma transacción de cada cambio de negocio.       | Interno (transacción de escritura, DAS §8)                                                 | N/A (interno)                                  | H2 real            |
| RF-FOR-12 | Proveer microcomponentes React moldeables  | Entregar microcomponentes autocontenidos (renderizador dinámico, listados y editor) personalizables por CSS, widgets y textos, más una app de desarrollo local con MSW.       | Paquete `@pingeso-civil/formularios-ui` (`apps/web/src/componentes/<kebab-case>/`)         | Contexto por props (sesión en cookie)          | H1 mocks / H2 real |

## Reglas de negocio (RN-FOR)

Toda regla `RN-FOR` debe tener al menos una fila en
[`docs/modos-de-falla.md`](modos-de-falla.md) indicando qué impide su violación
y cómo se verifica.

| ID        | Regla                                                       | Descripción                                                                                                                                                                       |
| --------- | ----------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| RN-FOR-01 | Inmutabilidad de versiones publicadas y no borrado físico   | Una versión en estado `publicada` no puede editarse ni borrarse físicamente. Los formularios con versiones o respuestas no se borran físicamente: se archivan lógicamente.        |
| RN-FOR-02 | Un único borrador de versión por formulario                 | Un formulario puede tener como máximo una versión en estado `borrador` a la vez; intentar crear un segundo borrador simultáneo se rechaza con HTTP 409.                           |
| RN-FOR-03 | Numeración correlativa y única de versiones                 | Las versiones de un formulario llevan un número entero `>= 1` único y correlativo por formulario; publicaciones concurrentes nunca deben generar números duplicados.              |
| RN-FOR-04 | Inmutabilidad de respuestas enviadas y no borrado físico    | Una respuesta en estado `enviada` fija `datos`, `enviada_en` y `hash_contenido`, y no puede editarse, reenviarse ni borrarse físicamente (historial protegido por `RESTRICT`).    |
| RN-FOR-05 | Respuestas solo sobre versión publicada y formulario activo | No se permite iniciar, editar borrador ni enviar respuestas sobre una versión en estado `borrador` ni sobre un formulario en estado `archivado` (rechazo HTTP 409).               |
| RN-FOR-06 | Validación obligatoria en el servidor al enviar             | Al enviar una respuesta, el servidor valida `datos` contra el JSON Schema 2020-12 de la versión (incluidos formatos custom como `rut`); cualquier error responde HTTP 422.        |
| RN-FOR-07 | Envío único y atómico de respuesta                          | La transición `borrador -> enviada` se ejecuta una sola vez de forma atómica; dos envíos concurrentes sobre la misma respuesta permiten solo uno y rechazan el resto con 409.     |
| RN-FOR-08 | Seguridad, validez y límites del esquema de formulario      | Todo `esquema` y `esquema_ui` debe ser JSON Schema 2020-12 válido, sin referencias externas (`$ref` remoto), dentro de límites explícitos de tamaño y profundidad.                |
| RN-FOR-09 | Autenticación JWT, autorización por permiso y propiedad     | Salvo `/healthz` y `/readyz`, toda operación exige JWT RS256 válido (vía JWKS) y el permiso de `x-permiso`. Sin `respuestas:administrar`, cada usuario accede solo a su `sub`.    |
| RN-FOR-10 | Auditoría transaccional y privacidad de datos sensibles     | Toda mutación de negocio escribe en `auditoria` en la misma transacción. Ni logs ni `auditoria` registran contraseñas, tokens, cookies ni el contenido de respuestas (solo IDs).  |
| RN-FOR-11 | Neutralidad de dominio y dueño único por dato               | El motor no incorpora lógica específica de un producto cliente, no accede a bases de datos de otros servicios y referencia entidades externas solo con IDs opacos (`UUID`/texto). |

## Requerimientos no funcionales (RNF-FOR)

Derivados del [DAS](referencia/das.md) y [Hitos 1 y 2](referencia/hitos-1-y-2.md):

| ID         | Categoría                     | Requerimiento                                                                                                                                                                                                                                                                                                                                                              | Referencia     |
| ---------- | ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| RNF-FOR-01 | Contrato API-first            | OpenAPI 3.1 generado con Huma (`pnpm codegen`) y commiteado; rutas en español plural `/api/...`; JSON `snake_case`; UUIDs; errores RFC 9457; `/api/docs` solo fuera de `prod`; compatibilidad verificada por `oasdiff` contra `main`.                                                                                                                                      | DAS §2, §6, §9 |
| RNF-FOR-02 | Arquitectura y mantenibilidad | Capas estrictas `Handler (internal/api) -> Service (internal/service) -> Store (internal/store)` sin saltos ni ciclos; microcomponentes React independientes sin estado global.                                                                                                                                                                                            | DAS §2, §3     |
| RNF-FOR-03 | Seguridad y configuración     | Variables de entorno validadas al arrancar (`API_PORT`, `API_ENV`, `DATABASE_URL`, `AUTH_JWKS_URL`, `FRONTEND_URL`, `TZ`); cero secretos en repo/imagen/`VITE_*`; sesión en cookie `HttpOnly`/`Secure`/`SameSite=Lax` (prohibidos `localStorage`/`sessionStorage`); timeouts explícitos de 5 s en clientes HTTP; `Idempotency-Key` en mutaciones hacia servicios externos. | DAS §2, §5, §6 |
| RNF-FOR-04 | Persistencia PostgreSQL       | PostgreSQL 17.11 con `pgx/v5`, migraciones goose inmutables ejecutadas por contenedor `migrate` (la API nunca migra), queries sqlc sin editar a mano, `timestamptz`, `uuid` y `jsonb`.                                                                                                                                                                                     | DAS §3, §5     |
| RNF-FOR-05 | Observabilidad y salud        | Logs `log/slog` JSON a stdout con `servicio`, `version`, `request_id`, `metodo`, `ruta`, `estado`, `duracion_ms`; propagación y sanitización de `X-Request-Id`; `/healthz` y `/readyz` sin auth.                                                                                                                                                                           | DAS §8         |
| RNF-FOR-06 | Calidad, seguridad y pruebas  | Jobs obligatorios `calidad`, `seguridad` e `imagen`; umbrales Hito 2: SonarQube (código nuevo: 0 bugs/vulnerabilidades, 100 % hotspots, <3 % duplicación, ≥60 % cobertura vía e2e), Trivy (0 HIGH/CRITICAL con fix), Gitleaks (0 hallazgos), Hadolint y Dockle; e2e con PostgreSQL real, golden determinista y pruebas con `TZ=UTC` y `TZ=Pacific/Auckland`.               | DAS §6, §7, §9 |
| RNF-FOR-07 | Contenedorización             | Dockerfiles multi-stage con imágenes base fijadas por digest, runtime `distroless:nonroot`, puerto interno 8080, Compose con `postgres -> migrate -> backend -> frontend` en puertos `51N0`/`51N1`.                                                                                                                                                                        | DAS §3, §4     |
| RNF-FOR-08 | Moldeabilidad y accesibilidad | Tema 100 % por variables CSS sin colores/radios/fuentes fijos, `className` y clases por parte, registro de widgets, textos por props, teclado completo, `label`/`aria-describedby`/`aria-invalid` y contraste AA.                                                                                                                                                          | DAS §3, AGENTS |
