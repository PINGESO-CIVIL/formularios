# Preguntas abiertas

Dudas sobre integración con otros microservicios de la LPS o sobre el alcance
de este MS. Cada `TODO(integracion)` del código debe tener una entrada aquí.

| # | Tema | Pregunta | Afecta a | Estado |
| --- | --- | --- | --- | --- |
| 1 | Alcance | ¿Qué responsabilidades quedan en `ms-formularios` y cuáles en los demás MS (usuarios/roles, archivos, notificaciones, trazabilidad, reportes)? | Todo | Abierta |
| 2 | Identidad | Formato real del token del MS de usuarios/roles (algoritmo, claims, cómo viaja el tenant). | `ports.Authenticator` | Abierta |
| 3 | Permisos | Nombres reales de permisos (hoy provisionales: `forms:read`, `forms:design`, `forms:fill`, `forms:review`, `forms:assign`, `forms:export`). | Backend, frontend | Abierta |
| 4 | Archivos | API del MS de archivos para verificar/subir adjuntos y firmas. | `ports.FileStore` | Abierta |
| 5 | Eventos | Broker elegido para la LPS y formato definitivo del sobre de eventos. | `ports.EventPublisher`, `contracts/events/` | Abierta |
| 6 | Reportes | ¿Exportación y reportes viven aquí o en el MS de reportes? | Alcance | Abierta |
