# Modos de falla — formularios

Cada falla posible del servicio, qué la impide y cómo se verifica (DAS §9). Se
escribe antes de programar cada parte y se actualiza cuando aparece una falla nueva.

Estado:

- **P**: prevención implementada y probada.
- **I**: implementada, falta la prueba indicada.
- **D**: diseñada, falta implementarla.
- **A**: riesgo aceptado y documentado.

## 1. Disponibilidad y arranque

| ID    | Falla                                            | Qué la impide                                                                                     | Cómo se verifica                                                                | Estado |
| ----- | ------------------------------------------------ | ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | ------ |
| MF-01 | PostgreSQL caído o sin red                       | `/readyz` responde 503 (problem+json, sin detalle interno) si el ping no contesta en 2 s          | `TestReadyzConBaseCaida`; prueba manual deteniendo el contenedor                | P      |
| MF-02 | Falta una variable de entorno o es inválida      | La configuración se valida al arrancar y el proceso termina con un error que la nombra            | `TestCargarFallaSiFaltanVariables`, `TestLevantarServidorFallaSinConfiguracion` | P      |
| MF-03 | La API arranca antes que las migraciones         | Compose: `backend` espera `service_completed_successfully` de `migrate`; la API nunca migra       | `docker compose up -d --build` desde cero                                       | P      |
| MF-04 | Consulta lenta deja la petición colgada          | Timeout en cada llamada a la base y `ReadHeaderTimeout` de 5 s en el servidor                     | e2e con consulta bloqueada (Hito 2)                                             | D      |
| MF-05 | Documentación interactiva expuesta en producción | `/api/docs` y el spec solo se publican con `API_ENV` distinto de `prod`                           | `TestDocsSoloFueraDeProduccion`                                                 | P      |
| MF-06 | usuarios-roles caído: no se puede leer el JWKS   | El JWKS se cachea (keyfunc); sin llaves válidas se responde 401 claro, nunca se acepta sin firmar | Prueba con JWKS inaccesible (Hito 1)                                            | D      |

## 2. Autenticación y autorización

| ID    | Falla                                             | Qué la impide                                                                                         | Cómo se verifica                                     | Estado |
| ----- | ------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | ------ |
| MF-10 | Token ausente, alterado, vencido o de otro emisor | Validación RS256 con JWKS, `iss=pingeso-identidad`, `aud=pingeso`, `exp` obligatorio, 30 s de desfase | Pruebas del middleware con tokens inválidos (Hito 1) | D      |
| MF-11 | Confusión de algoritmo (`alg: none`, HS256)       | El validador acepta solo `RS256`                                                                      | Prueba con `alg none` y HS256 (Hito 1)               | D      |
| MF-20 | Operación sin permiso declarado                   | Cada operación declara `x-permiso`; una prueba recorre el contrato                                    | `TestCadaOperacionDeclaraSuPermiso`                  | P      |
| MF-21 | Usuario sin permiso accede a un recurso           | Middleware compara `x-permiso` con el claim `permisos`; responde 403                                  | `acceso_test.go` (matriz por permiso, Hito 2)        | D      |
| MF-22 | Un usuario lee o modifica respuestas de otro      | Sin `respuestas:administrar`, el service filtra y valida por `creado_por` = `sub`                     | e2e de acceso cruzado (Hito 2)                       | D      |

## 3. Reglas del dominio

| ID    | Falla                                                      | Qué la impide                                                                                     | Cómo se verifica                       | Estado |
| ----- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------- | -------------------------------------- | ------ |
| MF-30 | Se modifica una versión ya publicada                       | El service responde 409; en la base, restricción de estado y sin `UPDATE` de versiones publicadas | Prueba del service y e2e               | D      |
| MF-31 | Se modifica o reenvía una respuesta ya enviada             | El service responde 409; `hash_contenido` fijo al enviar                                          | Prueba del service y e2e               | D      |
| MF-32 | Dos envíos simultáneos de la misma respuesta               | `UPDATE ... WHERE estado = 'borrador'` en una transacción: solo uno gana, el otro recibe 409      | e2e concurrente con golden (Hito 2)    | D      |
| MF-33 | Dos borradores de versión simultáneos del mismo formulario | Índice único parcial `versiones_formulario_un_borrador`                                           | e2e concurrente (Hito 2)               | D      |
| MF-34 | Respuesta sobre una versión en borrador o archivada        | El service solo acepta versiones `publicada` de formularios `activo`                              | Prueba del service                     | D      |
| MF-35 | Datos que no cumplen el esquema del formulario             | Validación JSON Schema 2020-12 en el servidor al enviar; 422 con error por campo                  | Prueba del service con datos inválidos | D      |
| MF-36 | Esquema de formulario malicioso (`$ref` externo, enorme)   | Solo referencias internas, límite de tamaño y profundidad al publicar                             | Prueba del service                     | D      |
| MF-37 | Borrado físico de historial                                | `ON DELETE RESTRICT` y archivado lógico                                                           | Migración + e2e                        | D      |

## 4. Información sensible

| ID    | Falla                                             | Qué la impide                                                                                    | Cómo se verifica                                    | Estado |
| ----- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------ | --------------------------------------------------- | ------ |
| MF-40 | Contenido de respuestas, tokens o cookies en logs | El middleware solo registra `request_id`, método, ruta, estado y duración; la auditoría solo IDs | Revisión de código y prueba sobre la salida de logs | I      |
| MF-41 | Errores que exponen trazas o SQL                  | RFC 9457 con mensajes genéricos                                                                  | `TestReadyzConBaseCaida` revisa que no haya detalle | P      |
| MF-42 | Secretos en el repositorio o la imagen            | Gitleaks (pre-commit y CI), `.env` y `*.pem` ignorados, `.dockerignore`                          | Job `seguridad`                                     | I      |
| MF-43 | Sesión robada desde el navegador                  | Cookie `HttpOnly`; `localStorage`/`sessionStorage` prohibidos en el front                        | Revisión y regla de lint (Hito 1)                   | D      |
| MF-44 | `X-Request-Id` malicioso inyectado en los logs    | Solo se acepta `[A-Za-z0-9._-]{1,128}`; si no, se genera uno nuevo                               | `TestRequestIDSePropagaOSeGenera`                   | P      |
