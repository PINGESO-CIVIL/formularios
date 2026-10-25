# AGENTS.md — ms-formularios

Guía para agentes de código (y personas) que trabajan en este repositorio.
Las carpetas `backend/` y `frontend/` tienen su propio `AGENTS.md` con reglas
específicas; cuando trabajes dentro de una de ellas, ese archivo manda sobre este.

## Qué es este proyecto

`ms-formularios` es el microservicio de **Formularios** de una Línea de
Productos de Software (LPS) de gestión de trazabilidad/inventario. Permite que
cada producto de la línea defina formularios (contratos, checklists, fichas,
encuestas, consentimientos), los publique en versiones, reciba respuestas
validadas y firmadas, y emita un evento por cada acción.

Productos que lo usarán (cada uno es un *tenant* con su propia configuración):

| Producto | Uso principal de formularios |
| --- | --- |
| InverFit | Contrato de servicio y aceptación de políticas con firma |
| MAD360 | Ficha de talento, derechos de imagen, solicitudes de casting |
| Gatica y Gatica | Checklists diarios de camión, charlas de seguridad con firmas |
| SkinLaser | Ficha de paciente, consentimiento, encuestas de satisfacción |
| TO USACH | Consentimiento informado, autoevaluación |

Documento de requerimientos (fuente de verdad funcional): ver `docs/requerimientos.md`
o el doc compartido del equipo. Los IDs `RF-FOR-xx`, `RNF-FOR-xx` y `RN-FOR-xx`
vienen de ahí; úsalos en commits, PRs y comentarios de pruebas.

## Estado actual: fase 0 (base independiente)

Los demás microservicios de la LPS (usuarios/roles, archivos, notificaciones,
trazabilidad, reportes) **todavía no existen o no tienen contrato definido**.
Por eso, en esta fase:

- **Sí** construimos lo que no depende de ellos: esqueleto, dominio de
  plantillas/versiones/respuestas, validación por JSON Schema, persistencia,
  API REST, renderer de formularios en Vue, pruebas, CI.
- **No** inventamos integraciones. Toda dependencia externa se modela como una
  interfaz (*puerto*) con una implementación local de desarrollo (*stub*).
  Cuando el otro MS exista, solo se agrega un adaptador nuevo.

| Dependencia externa | Puerto (backend) | Implementación en fase 0 |
| --- | --- | --- |
| MS usuarios/roles (identidad, tenant, permisos) | `ports.Authenticator` | `devauth`: JWT firmado con clave local (HS256) o headers `X-Dev-*` solo en modo `dev` |
| MS archivos (fotos, PDF, firmas) | `ports.FileStore` | Sistema de archivos local o MinIO del `docker-compose` |
| Bus de eventos / MS trazabilidad | `ports.EventPublisher` | Tabla `outbox_events` + publicador que escribe en log (`logpublisher`) |
| MS notificaciones | (ninguno) | No se llama; solo se emiten eventos |
| Entidades de negocio (cliente, camión, talento, paciente) | (ninguno) | Se referencian como `subject_type` + `subject_id` opacos; nunca se validan contra otro servicio |

Si una tarea requiere saber cómo se comporta otro MS (formato real de su token,
nombre real de un permiso, broker elegido), **no lo supongas en el código de
dominio**: déjalo en el adaptador, marca `// TODO(integracion): ...` y anótalo en
`docs/preguntas-abiertas.md`.

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
│   ├── requerimientos.md
│   ├── preguntas-abiertas.md
│   └── adr/                  # decisiones de arquitectura (ADR-0001-....md)
├── docker-compose.yml        # postgres, minio y, más adelante, el bus
└── .github/workflows/        # CI
```

## Reglas globales

1. **Contrato primero.** Cualquier cambio de API se hace primero en
   `contracts/openapi.yaml`; luego backend y frontend. Un PR que cambia un
   endpoint sin tocar el contrato se rechaza.
2. **Sin lógica de un producto en el núcleo.** Nada de `if tenant == "gatica"`.
   Las diferencias entre productos se expresan como *features* en
   `tenant_features` (ver requerimientos, sección de variabilidad).
3. **Multi-tenant siempre.** Toda tabla de negocio tiene `tenant_id` y toda
   consulta filtra por él. El `tenant_id` sale de la identidad autenticada,
   nunca del cuerpo de la petición.
4. **Inmutabilidad.** Una versión publicada y una respuesta enviada o firmada
   no se modifican ni se borran físicamente (RN-FOR-01, RN-FOR-04).
5. **Sin bitácora propia.** El historial lo construye trazabilidad a partir de
   los eventos. No crees tablas de auditoría; emite eventos.
6. **Datos sensibles.** No registres en logs el contenido de respuestas, RUT,
   firmas ni tokens. Solo IDs.
7. **Idioma.** Código, identificadores y nombres de tablas en inglés. Textos de
   UI, mensajes de error al usuario y documentación en español. Los nombres de
   eventos siguen el formato de la LPS en español (`respuesta.enviada`).
8. **Pequeño y verificable.** Un PR = un requerimiento o una pieza de
   infraestructura, con pruebas. Si algo es grande, divídelo.

## Convenciones de trabajo

- Ramas: `feat/RF-FOR-10-iniciar-respuesta`, `fix/...`, `chore/...`.
- Commits: Conventional Commits, con el ID si aplica:
  `feat(submissions): iniciar respuesta sobre versión publicada (RF-FOR-10)`.
- Decisiones que afecten a más de un módulo → ADR corto en `docs/adr/`.
- Antes de dar una tarea por terminada, corre las verificaciones de la carpeta
  donde trabajaste (`make check` en backend, `pnpm check` en frontend).

## Definición de terminado

- Compila, lint sin errores, pruebas en verde.
- Contrato OpenAPI actualizado si cambió la API.
- Pruebas que cubren el caso feliz y al menos un caso de error por regla.
- Sin `TODO` nuevos salvo `TODO(integracion)` registrados en preguntas abiertas.
- Migraciones reversibles (archivo `down` incluido).
