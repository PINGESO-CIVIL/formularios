# Preguntas abiertas

Dudas sobre integración con otros paquetes de trabajo (PT) de la LPS o sobre
el alcance de este MS (PT4). Cada `TODO(integracion)` del código debe tener una
entrada aquí.

| # | Tema | Pregunta | Con quién | Afecta a | Estado |
| --- | --- | --- | --- | --- | --- |
| 1 | Base de datos | ¿Se confirma MongoDB como base de datos única (lineamientos) o se adopta PostgreSQL? | Equipo docente / Equipo PT4 | Backend, `docker-compose`, CI | Resuelta: se adopta PostgreSQL + `JSONB` (ver `ADR-0003`) |
| 2 | Identidad | Formato real del JWT de PT6: algoritmo, claims, cómo viaja el tenant, cómo se obtiene la clave pública. | PT6 Identity Provider | `ports.Authenticator` | Abierta |
| 3 | Permisos | Nombres reales de roles/permisos (hoy provisionales: `forms:read`, `forms:design`, `forms:fill`, `forms:review`). | PT6 Identity Provider | Backend, frontend | Abierta |
| 4 | Archivos | API de PT7 para subir/verificar adjuntos y firmas; qué identificador devuelve. | PT7 Bóveda Documental | `ports.FileStore` | Abierta |
| 5 | Auditoría | Formato y transporte de los logs que recibe PT8 (HTTP, broker, campos obligatorios). | PT8 Central Auditoría | `ports.EventPublisher`, `contracts/events/` | Abierta |
| 6 | Límite con PT3 | PT3 "crea, versiona y almacena estructuras de documentos". ¿Dónde termina una plantilla de documento y empieza una definición de formulario? ¿Debemos consumir algo de PT3? | PT3 Gestor de Plantillas | Alcance | Abierta |
| 7 | Reportes | ¿PT1 generará PDF/Excel de respuestas de formularios? ¿Qué endpoint/forma de datos necesita de nosotros? | PT1 Motor de Reportes | Contrato | Abierta |
| 8 | Calidad | Servidor, token y umbrales de SonarQube del curso. | Equipo docente | CI | Abierta |
