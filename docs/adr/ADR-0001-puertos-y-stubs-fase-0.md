# ADR-0001: Puertos y stubs para dependencias externas en fase 0

- Estado: aceptado
- Fecha: 2026-10-25

## Contexto

Los demás microservicios de la LPS (usuarios/roles, archivos, notificaciones,
trazabilidad, reportes) no existen todavía o no tienen contrato definido, y aún
no está claro qué responsabilidades le corresponden a `ms-formularios`.

## Decisión

- Arquitectura hexagonal en el backend: `domain` ← `app` ← `adapters`, con
  `ports` como interfaces hacia afuera y `cmd` como único punto de wiring.
- Cada dependencia externa se modela como un puerto con un stub local
  (`devauth`, `localfs`, `logpublisher`).
- Los eventos se escriben en una tabla outbox en la misma transacción.
- El frontend trabaja contra `contracts/openapi.yaml` con MSW, sin backend.

## Consecuencias

- Se puede avanzar sin conocer los otros MS; integrar es agregar un adaptador.
- Las decisiones que dependen de otros MS quedan en `docs/preguntas-abiertas.md`.
- Hay que mantener los stubs fuera de producción (`devauth` solo con `APP_ENV=dev`).
