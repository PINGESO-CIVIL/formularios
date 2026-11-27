# Formato del esquema de formulario

> **Estado: borrador.** Este documento se completa junto con RF-FOR-01/04
> (plantillas y versiones) y el paquete `schema` del backend.

Un formulario versionado se describe con dos documentos:

| Documento | Estándar | Responsabilidad |
| --- | --- | --- |
| `schema` | JSON Schema 2020-12 | Estructura y validación de los datos (la que vale es la del servidor) |
| `uiSchema` | Propio (este documento) | Orden, secciones, widgets, condiciones de visibilidad |

## Formatos personalizados

| Formato | Descripción |
| --- | --- |
| `rut` | RUT chileno con dígito verificador |

## Pendiente de definir

- Estructura del `uiSchema` (secciones, orden, widgets, condiciones).
- Catálogo de tipos de campo de fase 0 y su mapeo a JSON Schema.
- Ejemplos por producto (InverFit, MAD360, Gatica, SkinLaser, TO USACH).
