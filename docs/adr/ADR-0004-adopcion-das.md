# ADR-0004: Adopción del DAS como norma técnica

- Estado: aceptado
- Fecha: 2026-11-24
- Reemplaza: ADR-0001, ADR-0002 y ADR-0003 en lo que se opongan a este documento

## Contexto

El curso publicó el **DAS — Desarrollo de Microservicios**
(`docs/referencia/das.md`) y el documento de **Hitos 1 y 2**
(`docs/referencia/hitos-1-y-2.md`). Ambos son posteriores a los lineamientos
iniciales y fijan, con versiones exactas, el stack, la estructura del repositorio,
el pipeline y los controles de seguridad. Los dos hitos se entregan el 4 de
diciembre, en el repositorio de la organización `PINGESO-CIVIL`.

La base actual del repo (`backend/` con arquitectura hexagonal, `frontend/` con
Vue, migraciones golang-migrate, workflows separados) no coincide con la
plantilla que exige el DAS.

## Decisión

1. El DAS es la norma técnica; donde los lineamientos o los ADR anteriores lo
   contradicen, gana el DAS.
2. **Estructura:** `apps/api` (Go) y `apps/web` (React), igual a la plantilla.
   Se retiran `backend/` y `frontend/`. Como referencia de la plantilla se usa
   el repositorio `PINGESO-CIVIL/usuarios-roles`.
3. **API:** Go 1.26 + Huma v2. El contrato OpenAPI 3.1 se **genera** desde los
   structs con `pnpm codegen`; no se escribe a mano.
4. **Capas:** `internal/api` (handler) → `internal/service` → `internal/store`.
5. **Datos:** PostgreSQL 17.11 con migraciones **goose** y queries **sqlc**.
   Se mantiene la idea del `ADR-0003` (estructuras dinámicas en `JSONB`), pero
   las migraciones se reescriben para goose.
6. **Frontend:** React 19 como microcomponentes (DAS §3). Se retira el monorepo Vue.
7. **Errores:** RFC 9457 (`problem+json`). Se descarta el `ErrorEstandar` del
   swagger de ejemplo.
8. **Rutas:** `/api/<recurso>` en español y plural, ids UUID. Sin `/v1` hasta que
   exista un cambio incompatible.
9. **Autenticación:** JWT RS256 de `usuarios-roles`, validado con `AUTH_JWKS_URL`;
   autorización por permiso declarado en `x-permiso`.
10. **Sin multi-tenant:** el JWT de `usuarios-roles` no incluye tenant y cada
    cliente despliega su propia instancia. Se eliminan `tenant_id` y
    `tenant_features`.
11. **Auditoría propia:** tabla `auditoria` en la base del servicio (DAS §8), en
    vez de depender de la Central de Auditoría.
12. **Idioma:** identificadores y comentarios en español (DAS §9).
13. **Puertos locales:** se asume N = 4 (PT4): web `5140`, api `5141`, hasta
    confirmar con el documento de componentes.

## Desviaciones registradas

- **Toolchain Go:** el DAS fija `go1.26.5`. Se usa `go1.26.11` (el mismo de
  `usuarios-roles`), porque las versiones anteriores tienen vulnerabilidades de
  la biblioteca estándar que bloquean govulncheck y Trivy. Requiere visto bueno
  del jefe de proyecto.

## Consecuencias

- Se reescribe la base del servicio, pero se conserva el diseño funcional
  (formularios, versiones, respuestas, `esquema` + `esquema_ui`).
- El trabajo de PostgreSQL y CI de las tandas anteriores se migra a la nueva
  estructura: modelo relacional a goose y SonarQube/Trivy al `ci.yml` del DAS.
- El nombre del servicio, el número N y las dependencias oficiales quedan
  pendientes del documento de componentes (`docs/preguntas-abiertas.md`).
