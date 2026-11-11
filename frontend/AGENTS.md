# AGENTS.md — frontend (Vue)

Reglas para trabajar en el frontend de `ms-formularios`. Lee también el
`AGENTS.md` de la raíz: lineamientos del curso, contrato primero, sin lógica
por producto, textos de UI en español.

## Qué se construye aquí

No es una sola app: es un **conjunto de paquetes** que cualquier producto
derivado de la LPS podrá ensamblar para mostrar y diseñar formularios
(lineamiento: *frontends modulares*, vistas aisladas para fácil ensamblaje
final), más una app de administración que sirve de *playground*.

```
frontend/                     # monorepo pnpm workspaces
├── packages/
│   ├── forms-core/           # tipos TS del esquema, cliente API, validación espejo
│   ├── forms-renderer/       # <FormRenderer>: muestra y llena un formulario
│   └── forms-builder/        # <FormBuilder>: diseña plantillas (fase posterior)
└── apps/
    └── forms-admin/          # app Vite: plantillas, versiones, respuestas
```

Dependencias permitidas entre paquetes: `forms-renderer` → `forms-core`;
`forms-builder` → `forms-renderer`, `forms-core`; `forms-admin` → todos.
`forms-core` no depende de Vue.

## Stack

| Tema | Elección |
| --- | --- |
| Framework | **Vue 3.5** (lineamientos), Composition API con `<script setup lang="ts">` |
| Build | Vite; los paquetes se publican en modo librería (`vite build --lib`) |
| Lenguaje | TypeScript estricto (`strict: true`, sin `any` implícito) |
| Estado (solo en `forms-admin`) | Pinia |
| Rutas (solo en `forms-admin`) | Vue Router |
| Validación en cliente | Ajv (JSON Schema 2020-12) + formato `rut` |
| API mock | MSW (Mock Service Worker) |
| Pruebas | Vitest + Vue Test Utils; Playwright para e2e de `forms-admin` |
| Lint / formato | ESLint (flat config) + Prettier |
| Calidad / seguridad | SonarQube (cobertura de Vitest en `lcov`) y Trivy (dependencias pnpm) |
| Node / gestor | Node 22 LTS o superior, pnpm |

## Comandos

```bash
pnpm install
pnpm dev               # forms-admin con MSW activo (no requiere backend)
pnpm dev:api           # forms-admin contra backend local (VITE_API_URL)
pnpm test              # vitest en todos los paquetes
pnpm e2e               # playwright sobre forms-admin
pnpm lint
pnpm typecheck
pnpm build
pnpm gen:api           # regenera tipos desde ../contracts/openapi.yaml
pnpm check             # lint + typecheck + test: córrelo antes de terminar
```

Si un script no existe todavía en `package.json`, créalo como parte de tu tarea.

## Trabajar sin backend ni otros MS

El backend se construye en paralelo y los demás MS no existen todavía:

- **API:** los tipos se generan desde `contracts/openapi.yaml`
  (`openapi-typescript`). El cliente vive en `forms-core/api/` y recibe `baseUrl`
  y una función `getToken()` inyectadas; nunca lee `localStorage` por su cuenta.
- **Mocks:** cada endpoint tiene un handler MSW en `apps/forms-admin/src/mocks/`
  con datos de ejemplo **genéricos** (contrato con aceptación, checklist,
  ficha de registro, encuesta, consentimiento). Sin nombres de productos reales.
- **Identidad (PT6):** `forms-admin` usa un selector de "usuario dev" (tenant +
  permisos) que entrega un JWT de prueba quemado, igual al mock del backend.
  Está detrás de `import.meta.env.DEV` y no se incluye en el build de producción.
  El login real es de PT6; aquí no se construye.
- **Archivos:** `FileField` y `SignatureField` reciben un `uploader` inyectado
  (`(file: Blob) => Promise<{ fileId: string }>`). En fase 1 se usa un uploader
  falso; el real será el de PT7 Bóveda Documental.

## Renderer: reglas de diseño

- Entrada: `schema` (JSON Schema), `uiSchema` (orden, secciones, widgets,
  condiciones) y `v-model` con los datos. Salida: eventos `submit`,
  `save-draft`, `change`. El renderer **no** llama a la API: eso lo hace quien
  lo usa.
- **Registro de campos = punto de variabilidad de la LPS.** Cada tipo de campo
  se registra con `registerField(type, component)`. Un producto puede
  registrar o reemplazar campos sin tocar el paquete.
- Contrato de un componente de campo:
  `props: { modelValue, field, errors, disabled, readonly }`, `emits: ['update:modelValue', 'blur']`.
- Campos de fase 1: `text`, `textarea`, `number`, `date`, `time`, `select`,
  `multiselect`, `boolean`, `rut`, `email`, `phone`. Después: `file`,
  `signature`, `repeater`, `scale`.
- La validación del cliente es para la experiencia de uso; la que vale es la
  del servidor. Muestra los errores del `ErrorEstandar` (`errores[].campo`) en el
  campo correspondiente, y `mensaje` como error general.
- Sin librería de UI acoplada: estilos con variables CSS (`--forms-color-primary`,
  `--forms-radius`, `--forms-font`, …) definidas en `forms-renderer/src/theme.css`,
  para que cada producto aplique su marca.
- Mobile first: debe usarse cómodo en un celular (formularios llenados en
  terreno).

## Accesibilidad (WCAG 2.1 AA)

- Cada input con `<label for>`; errores enlazados con `aria-describedby` y `aria-invalid`.
- Todo operable con teclado, foco visible.
- Al fallar un envío, el foco va al primer campo con error.
- Contraste mínimo 4.5:1 con el tema por defecto.

## Convenciones

- Componentes en `PascalCase.vue`, composables `useAlgo.ts`, un componente por archivo.
- Textos de UI en español, centralizados en `forms-renderer/src/i18n/es.ts`.
- Nada de lógica específica de un producto en los paquetes. Si un producto
  necesita algo distinto, se resuelve con un campo registrado, una opción del
  `uiSchema` o una feature del tenant.
- Pruebas de componentes por comportamiento (lo que ve y hace el usuario), no
  por detalles internos.

## Qué construir ahora: Hito 1 (13 de noviembre), en este orden

1. Monorepo pnpm con los tres paquetes y `forms-admin`, TS estricto, ESLint, Prettier, Vitest, CI con SonarQube y Trivy.
2. `forms-core`: tipos del esquema de formulario (según `contracts/form-schema.md`), cliente API tipado desde OpenAPI, validador Ajv con formato `rut`.
3. `forms-renderer`: `FormRenderer`, registro de campos, campos de fase 1, mostrar errores del servidor, tema por variables CSS.
4. `forms-admin` con MSW: listar plantillas, ver versiones, vista previa de una versión con el renderer, llenar y enviar una respuesta, listar respuestas.
5. Editor de esquema simple (JSON con vista previa en vivo) antes del builder visual.

## Qué NO hacer en esta fase

- No construyas el `FormBuilder` visual de arrastrar y soltar todavía.
- No implementes login real, manejo de sesión ni roles: solo el selector dev.
- No llames a otros microservicios (PT6, PT7, PT8…) ni supongas sus URLs.
- No generes PDF/Excel (PT1) ni envíes correos (PT10) desde el frontend.
- No agregues librerías de UI (Vuetify, PrimeVue, etc.) a los paquetes.
