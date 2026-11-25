# AGENTS.md — apps/web (React, microcomponentes)

Reglas específicas del frontend. Aplican también las del `AGENTS.md` de la raíz y
el DAS (`docs/referencia/das.md`), que mandan sobre este archivo.

## Qué se construye aquí

**Microcomponentes** React que cada grupo cliente compone en su propio producto
(DAS §3), más una **app de desarrollo** (`src/app/`) que los monta con mocks MSW y
que no se publica. Se publicarán como `@pingeso-civil/formularios-ui`.

## Prioridad del equipo: visual moldeable

El servicio debe servir a productos muy distintos, así que **todo lo visual es
reemplazable sin tocar el paquete**:

1. **Tema por variables CSS** en `src/estilos/tema.css` (`--formularios-primario`,
   `--formularios-radio`, `--formularios-fuente`…). Tailwind las expone como clases
   (`bg-primario`, `rounded-formularios`, `font-formularios`). Un producto redefine
   las variables en su hoja y el componente adopta su marca.
2. **Nunca colores, radios ni fuentes fijos** en un componente: solo clases del tema.
3. Cada microcomponente acepta `className` para su contenedor y, cuando aplique,
   `clases` por parte (`{ campo, etiqueta, error, boton }`).
4. **Registro de widgets** en el renderizador: un producto puede registrar o
   reemplazar el componente de un tipo de campo (`text`, `select`, `rut`…).
5. Textos visibles en un objeto de textos reemplazable por props (por defecto en español).
6. Componentes UI base generados con shadcn (`src/components/ui/`, no se editan);
   las variantes van en un wrapper en `src/components/`.

## Estructura de un microcomponente (DAS §3)

```
src/componentes/<nombre-kebab>/
├── <nombre>.tsx       # vista, sin lógica de negocio
├── <nombre>.api.ts    # único lugar con fetch (openapi-fetch tipado)
├── use-<nombre>.ts    # estado (react-query / react-hook-form)
└── <nombre>.test.tsx
```

- Un microcomponente **no importa a otro** ni depende de un layout, ruta o estado global.
- Recibe por props la **URL base** de la API y lo que necesite del contexto
  (usuario, permisos).
- Tipos de la API: `src/api/esquema.d.ts`, generado con `pnpm codegen`. No se edita.

## Reglas

- `localStorage` y `sessionStorage` **prohibidos**. La sesión viaja en cookie.
- `fetch` solo en archivos `*.api.ts`.
- Ningún secreto en variables `VITE_*`.
- Dependencias con versión exacta; las compartidas van en el `catalog` de
  `pnpm-workspace.yaml`.
- Prettier sin punto y coma, ancho 90; oxlint sin advertencias.
- Accesibilidad: `label` asociado a cada campo, errores con `aria-describedby` y
  `aria-invalid`, todo operable con teclado, contraste AA con el tema por defecto.

## Comandos (desde la raíz)

```bash
pnpm --filter @pingeso-civil/formularios-ui dev   # app de desarrollo en :5140
pnpm typecheck
pnpm test
pnpm codegen
```

## Servidor web

`servidor/` es el spa-server en Go (módulo propio): sirve `dist/`, responde
`/healthz` y reenvía `/api` al backend (`BACKEND_URL`), para que el navegador vea
un solo origen.
