# syntax=docker/dockerfile:1.7
# Imagen de la app web (DAS §4): build de Vite, spa-server en Go y runtime distroless.

# 1. Build de la app con Vite
FROM --platform=$BUILDPLATFORM node:22.22.0-alpine3.23@sha256:e4bf2a82ad0a4037d28035ae71529873c069b13eb0455466ae0bc13363826e34 AS web
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/package.json apps/web/
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --ignore-scripts --filter @pingeso-civil/formularios-ui...
COPY apps/web/ apps/web/
RUN pnpm --filter @pingeso-civil/formularios-ui build

# 2. Compilación del spa-server
FROM --platform=$BUILDPLATFORM golang:1.26.11-alpine3.23@sha256:23f198931d4bd83b4f5a613bd75656257ae8bf832dcbc2db80e1050b5219ca2a AS servidor
WORKDIR /src
COPY apps/web/servidor/ ./
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/servidor .

# 3. Runtime mínimo, sin shell ni compilador
FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6
WORKDIR /app
COPY --from=servidor --chown=nonroot:nonroot /out/servidor /app/servidor
COPY --from=web --chown=nonroot:nonroot /src/apps/web/dist /srv
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/servidor"]
