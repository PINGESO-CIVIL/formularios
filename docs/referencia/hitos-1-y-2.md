# Hitos 1 y 2: entregables y evidencia

Cada grupo traspasa su repositorio a la organización del curso como público y desde ahí entrega el Hito 1 y el Hito 2. El detalle técnico de cada punto está en el DAS.

## 1. Traspaso a la organización

Todos los repositorios pasan a la organización <https://github.com/PINGESO-CIVIL> como públicos. Así todos ven el estado de todos, el pipeline corre sin límite de minutos y `main` se puede proteger como exige el DAS.

1. **Aceptar la invitación.** Cada integrante recibe un correo de GitHub en la dirección que el curso tiene registrada, con la invitación a `PINGESO-CIVIL`. Se acepta iniciando sesión con la cuenta de GitHub que usan en el proyecto. Quien es dueño del repositorio tiene que aceptarla sí o sí, porque sin eso no puede transferir. La invitación vence a los 7 días: si no llega, revisen spam y avisen al ayudante.
2. **Revisar antes de publicar.** El repositorio queda público con todo su historial. Revisen el historial completo con `gitleaks git --redact .` y confirmen que no hay datos confidenciales del cliente. Si aparece un secreto, se rota antes de seguir, porque borrarlo del historial no basta.
3. **Hacerlo público.** Quien es dueño del repositorio entra a Settings, sección Danger Zone, y cambia la visibilidad a pública.
4. **Transferir.** En la misma sección, con la invitación ya aceptada, el dueño elige Transfer ownership, selecciona `PINGESO-CIVIL` y deja como nombre el `<servicio>` del documento de componentes.
5. **Actualizar el remoto.** Cada integrante corre `git remote set-url origin https://github.com/PINGESO-CIVIL/<servicio>.git`.

Después del traspaso, los ayudantes asignan los permisos del equipo y configuran la protección de `main`. El grupo confirma lo siguiente:

- [ ] El repositorio aparece en `PINGESO-CIVIL` como público, con el nombre del servicio.
- [ ] Un push o PR nuevo deja los jobs `calidad`, `seguridad` e `imagen` en verde. Si fallan por los secrets de SonarQube, avisan al ayudante.
- [ ] Los enlaces del README al pipeline y a SonarQube apuntan a la nueva dirección.

## 2. Qué debe tener cada hito

Cada hito tiene un entregable que se verifica en el repositorio de la organización. Las fechas son comunes a ambas secciones.

| Hito | Fecha | Se aprueba con |
| --- | --- | --- |
| 1. Implementación inicial | Dec 4, 2026 | Contrato publicado y validado, endpoints respondiendo con mocks y revisión con nota de la arquitectura |
| 2. Refinamiento y calidad | Dec 4, 2026 | Lógica real sin mocks, persistencia en PostgreSQL y umbrales de Trivy y SonarQube superados |

### Hito 1: Implementación inicial

El servicio publica su contrato OpenAPI antes que su lógica, para que los grupos que dependen de él avancen en paralelo contra mocks. Todos los endpoints responden con datos simulados que cumplen el contrato, y el entorno local levanta completo con Docker Compose.

En este hito se revisa con nota la arquitectura: capas separadas, un solo dueño por dato, dependencias sin ciclos y los modos de falla del servicio escritos.

### Hito 2: Refinamiento y calidad

El servicio deja los mocks. La lógica es real, los datos persisten en su propia base PostgreSQL y las llamadas a otros servicios van contra sus APIs. Se suman la autorización por rol, la auditoría de cambios y al menos un e2e de dificultad media o alta contra PostgreSQL real.

El Hito 2 se aprueba cuando los escáneres pasan estos umbrales sobre el commit entregado:

| Escáner | Umbral para aprobar |
| --- | --- |
| SonarQube, código nuevo | 0 bugs y 0 vulnerabilidades nuevas, 100% de hotspots revisados, duplicación bajo 3% y cobertura de al menos 60% |
| Trivy, repositorio e imágenes | 0 hallazgos HIGH o CRITICAL con corrección disponible |
| Gitleaks | 0 hallazgos, sin excepciones |

Lo que verifica el pipeline se acepta con el run en verde. El resto lo revisa el ayudante en el repositorio.

## 3. Checklist del Hito 1

Es la misma lista de la plantilla del issue "Entrega Hito 1". El título de cada grupo indica la sección del DAS de donde sale.

### Contrato API · DAS §2

- [ ] Structs de request y response en `internal/api/`, con tags `json`, validaciones y ejemplos.
- [ ] `contracts/openapi.yaml` en OpenAPI 3.1 y los tipos TypeScript, generados con `pnpm codegen` y commiteados.
- [ ] Endpoints en español y plural, como `/api/usuarios`, con identificadores UUID.
- [ ] Toda entrada validada en el contrato: tipos, largos, máximos y patrones.
- [ ] Cada endpoint declara su permiso, por ejemplo `<recurso>:operar`.
- [ ] Contrato publicado para los grupos que dependen del servicio, según el documento de componentes de la sección.

### Endpoints con datos simulados · DAS §2, §4 y §8

- [ ] Todos los endpoints del contrato responden con datos simulados que cumplen el contrato.
- [ ] Documentación interactiva en `/api/docs` con `API_ENV=dev`.
- [ ] `GET /healthz` y `GET /readyz` responden.
- [ ] `docker compose up -d --build` levanta postgres, migrate, backend y frontend en los puertos `51N0` y `51N1` del servicio.

### Consumo de otros servicios · DAS §2

- [ ] Contrato de cada proveedor copiado en `contracts/externos/<servicio>.yaml`, con su cliente generado desde ahí.
- [ ] Mientras el proveedor no tenga lógica, el consumidor trabaja contra `prism mock` o handlers MSW.

### Arquitectura, revisión con nota · DAS §2 y §3

- [ ] Capas Handler, Service y Store separadas: ningún handler llama al store.
- [ ] Dueño único por dato: el servicio no escribe datos de otro ni se conecta a su base.
- [ ] Dependencias en un solo sentido, sin ciclos y coherentes con el documento de componentes.
- [ ] Estructura de repositorio igual a la de la plantilla.
- [ ] `docs/modos-de-falla.md` con cada falla posible, qué la impide y cómo se verifica.

### Pipeline · DAS §6 y §7

- [ ] Jobs `calidad`, `seguridad` e `imagen` en verde sobre el commit entregado.
- [ ] Paso de contrato en verde: `pnpm codegen` sin diferencias y oasdiff sin cambios incompatibles contra `main`.

## 4. Checklist del Hito 2

Es la lista de la plantilla "Entrega Hito 2". Todo lo del Hito 1 se sigue cumpliendo, y esta lista agrega lo propio del Hito 2.

### Lógica real · DAS §2 y §3

- [ ] Ningún endpoint devuelve datos simulados: cada handler llama a su service y el service a su store.
- [ ] Las llamadas a otros servicios van por `<SERVICIO>_URL`, reenviando el JWT y el `X-Request-Id`.
- [ ] Todo cliente HTTP tiene timeout explícito, 5 s por defecto, y un servicio caído devuelve un error claro.
- [ ] Una operación que cambia datos de otro servicio se confirma solo si ese servicio la confirma.

### Persistencia en PostgreSQL · DAS §5 y §8

- [ ] Esquema solo en migraciones goose de `db/migrations/`, sin editar migraciones ya desplegadas.
- [ ] Queries en `db/queries/` y `internal/repository/` generado con sqlc, sin cambios a mano.
- [ ] Montos en `bigint`, fechas en `timestamptz` e identificadores expuestos como UUID.
- [ ] El contenedor `migrate` corre antes que la API, y la API no migra al arrancar.
- [ ] Las operaciones que cambian datos de negocio quedan en una tabla de auditoría con quién, qué y cuándo.

### Seguridad de aplicación · DAS §5 y §6

- [ ] JWT validado contra `AUTH_JWKS_URL`, con la sesión en cookie `HttpOnly`, `Secure` y `SameSite=Lax`.
- [ ] Matriz de acceso por rol probada en `acceso_test.go`.
- [ ] `Idempotency-Key` en las operaciones que mueven cupos, inventario o dinero, y en las que cambian datos de otro servicio.
- [ ] Errores en RFC 9457, sin trazas, SQL ni rutas internas.
- [ ] Configuración validada al arrancar y ningún secreto en el repositorio, la imagen ni variables `VITE_*`.

### Pruebas e2e · DAS §9

- [ ] `docs/modos-de-falla.md` actualizado con las fallas de la lógica real.
- [ ] E2E que levanta el binario real de la API contra PostgreSQL real, con Testcontainers o `TEST_DATABASE_URL`.
- [ ] Al menos un escenario de dificultad media o alta: concurrencia, rechazos esperados u operaciones encadenadas.
- [ ] Reporte determinista comparado con un golden en `e2e/testdata/`.
- [ ] Evidencia de que el e2e falla al quitar a propósito la protección que prueba.
- [ ] Lógica con fechas probada también con `TZ=UTC` y `TZ=Pacific/Auckland`.

### Umbrales · DAS §6 y §7

- [ ] Quality Gate de SonarQube en verde sobre el commit entregado, con la cobertura medida desde los e2e.
- [ ] Jobs `seguridad` e `imagen` en verde.
- [ ] Cada excepción de Trivy en `.trivyignore.yaml` con CVE, motivo y fecha de revisión, aprobada por el ayudante.

### Observabilidad · DAS §8

- [ ] Logs JSON a stdout con `servicio`, `version`, `request_id`, `metodo`, `ruta`, `estado` y `duracion_ms`.
- [ ] Ningún log contiene contraseñas, tokens, cookies ni el cuerpo completo de una request.
- [ ] `/readyz` responde 200 solo si PostgreSQL contesta.

## 5. Formato de entrega

Cada hito se entrega como una release de GitHub en el repositorio de la organización, más un issue de entrega con la checklist. Ambos deben estar publicados antes del cierre del hito.

1. **Tag.** Crear el tag `hito-1` o `hito-2` sobre el commit de `main` que se entrega, con el pipeline en verde.
2. **Release.** Publicar una release desde ese tag con el `openapi.yaml` adjunto. En el Hito 2 se adjunta también el SBOM CycloneDX, generado con `trivy image --format cyclonedx`. Las notas de la release llevan los cambios y los enlaces de la tabla.
3. **Issue de entrega.** Abrir un issue con la plantilla "Entrega Hito 1" o "Entrega Hito 2", marcar lo que se cumple y enlazar la release. El ayudante revisa en ese issue y lo cierra cuando aprueba.

| Evidencia | Hito 1 | Hito 2 |
| --- | --- | --- |
| Commit evaluado | Tag `hito-1` en `main` | Tag `hito-2` en `main` |
| Pipeline | Run en verde de ese commit | Run en verde de ese commit |
| Contrato | `openapi.yaml` adjunto a la release | `openapi.yaml` adjunto y resultado de oasdiff contra `hito-1` |
| SBOM | No se pide | CycloneDX adjunto a la release |
| SonarQube | No se pide | Proyecto con el Quality Gate en verde para ese commit |
| Modos de falla | `docs/modos-de-falla.md` | `docs/modos-de-falla.md` actualizado |
| E2E | No se pide | Golden en `e2e/testdata/` y un run en rojo con la protección quitada |

Los tags `hito-*` quedan protegidos, así que no se pueden mover ni borrar. Lo que se cambie después del tag no cuenta para el hito.
