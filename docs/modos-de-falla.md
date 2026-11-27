# Modos de falla de la base del servicio

| Falla                                          | Comportamiento esperado                                           | Verificación                                                   |
| ---------------------------------------------- | ----------------------------------------------------------------- | -------------------------------------------------------------- |
| Configuración de PostgreSQL ausente o inválida | La API devuelve un error al arrancar.                             | Pruebas de configuración y del pool en `apps/api/internal`.    |
| PostgreSQL deja de responder                   | `/readyz` responde 503.                                           | Pruebas del router en `apps/api/internal/api`.                 |
| El contrato cambia                             | Se regeneran los tipos del cliente y se revisa la compatibilidad. | `pnpm codegen`, `pnpm typecheck` y comparación con main en CI. |
