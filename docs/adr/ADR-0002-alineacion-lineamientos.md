# ADR-0002: Alineación con los lineamientos técnicos del curso

- Estado: aceptado
- Fecha: 2026-11-11

## Contexto

El curso publicó "Lineamientos Técnicos y Plan Incremental"
(`docs/referencia/lineamientos.pdf`), que fija stack, arquitectura, paquetes de
trabajo (PT1–PT10) e hitos. La base del repo (fase 0) se había armado antes,
con supuestos propios: PostgreSQL, Go 1.26, MS externos con nombres inventados,
errores `problem+json`.

## Decisión

- Los lineamientos son la fuente de verdad técnica; el repo se ajusta a ellos.
- Somos **PT4 Motor de Formularios**. Dependencia oficial: PT6. Dependencias
  identificadas por el equipo: PT8 (auditoría, saliente) y PT7 (archivos).
  Límites explícitos con PT1, PT3 y PT10.
- Stack: OpenAPI 3.1.0, Go 1.23, Vue 3.5, MongoDB, SonarQube, Trivy.
- Arquitectura Handlers → Services → Repositories, mapeada sobre las carpetas
  existentes (ver `backend/AGENTS.md`).
- Contrato en el estilo del swagger de ejemplo del curso: rutas y campos en
  español, `snake_case`, `ErrorEstandar` (`codigo`, `mensaje`, `detalles`),
  IDs ObjectID.
- Para el Hito 1 los services son mocks con datos fijos; se reemplazan en el
  Hito 2 sin tocar los handlers.

## Consecuencias

- Reemplaza parcialmente a ADR-0001: se mantiene la idea de puertos + mocks,
  pero los stubs pasan a ser los "mocks requeridos" de los lineamientos
  (JWT quemado para PT6, consola para PT8, `file_id` falso para PT7) y se
  elimina el outbox sobre PostgreSQL.
- La elección de base de datos queda sujeta a confirmación del equipo docente
  (`docs/preguntas-abiertas.md` #1); mientras tanto el backend arranca sin BD.
