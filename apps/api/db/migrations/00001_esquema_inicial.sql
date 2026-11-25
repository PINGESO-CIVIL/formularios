-- Esquema inicial del servicio formularios (DAS §5).
-- Las estructuras dinámicas (esquema, esquema_ui, datos) van en JSONB.
-- Los usuarios se referencian por el UUID del claim sub del JWT de usuarios-roles:
-- este servicio no guarda nombres ni correos.
-- Sin ON DELETE CASCADE: versiones publicadas y respuestas enviadas no se borran.

-- +goose Up
CREATE TABLE formularios (
    id             uuid        PRIMARY KEY,
    nombre         text        NOT NULL CHECK (char_length(nombre) BETWEEN 1 AND 120),
    descripcion    text        NOT NULL DEFAULT '' CHECK (char_length(descripcion) <= 1000),
    categoria      text        NOT NULL DEFAULT '' CHECK (char_length(categoria) <= 60),
    etiquetas      text[]      NOT NULL DEFAULT '{}',
    estado         text        NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo', 'archivado')),
    creado_por     uuid        NOT NULL,
    creado_en      timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX formularios_estado_creado_en ON formularios (estado, creado_en DESC);

CREATE TABLE versiones_formulario (
    id            uuid        PRIMARY KEY,
    formulario_id uuid        NOT NULL REFERENCES formularios (id) ON DELETE RESTRICT,
    numero        integer     NOT NULL CHECK (numero >= 1),
    estado        text        NOT NULL DEFAULT 'borrador' CHECK (estado IN ('borrador', 'publicada')),
    esquema       jsonb       NOT NULL,
    esquema_ui    jsonb       NOT NULL DEFAULT '{}'::jsonb,
    notas         text        NOT NULL DEFAULT '' CHECK (char_length(notas) <= 500),
    creado_por    uuid        NOT NULL,
    creado_en     timestamptz NOT NULL DEFAULT now(),
    publicada_por uuid,
    publicada_en  timestamptz,
    CONSTRAINT versiones_formulario_numero_unico UNIQUE (formulario_id, numero),
    CONSTRAINT versiones_formulario_publicacion_completa CHECK (
        (estado = 'borrador' AND publicada_en IS NULL AND publicada_por IS NULL)
        OR (estado = 'publicada' AND publicada_en IS NOT NULL AND publicada_por IS NOT NULL)
    )
);

-- Solo puede existir un borrador por formulario.
CREATE UNIQUE INDEX versiones_formulario_un_borrador
    ON versiones_formulario (formulario_id) WHERE estado = 'borrador';

CREATE TABLE respuestas (
    id             uuid        PRIMARY KEY,
    formulario_id  uuid        NOT NULL REFERENCES formularios (id) ON DELETE RESTRICT,
    version_id     uuid        NOT NULL REFERENCES versiones_formulario (id) ON DELETE RESTRICT,
    estado         text        NOT NULL DEFAULT 'borrador' CHECK (estado IN ('borrador', 'enviada')),
    sujeto_tipo    text        CHECK (char_length(sujeto_tipo) <= 60),
    sujeto_id      text        CHECK (char_length(sujeto_id) <= 120),
    datos          jsonb       NOT NULL DEFAULT '{}'::jsonb,
    hash_contenido text,
    creado_por     uuid        NOT NULL,
    creado_en      timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now(),
    enviada_en     timestamptz,
    CONSTRAINT respuestas_envio_completo CHECK (
        (estado = 'borrador' AND enviada_en IS NULL AND hash_contenido IS NULL)
        OR (estado = 'enviada' AND enviada_en IS NOT NULL AND hash_contenido IS NOT NULL)
    )
);

CREATE INDEX respuestas_formulario_creado_en ON respuestas (formulario_id, creado_en DESC);
CREATE INDEX respuestas_sujeto ON respuestas (sujeto_tipo, sujeto_id)
    WHERE sujeto_tipo IS NOT NULL AND sujeto_id IS NOT NULL;

-- Auditoría propia (DAS §8): quién, qué y cuándo. Nunca guarda el contenido de las
-- respuestas, solo identificadores.
CREATE TABLE auditoria (
    id          bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_id    uuid        NOT NULL,
    accion      text        NOT NULL,
    recurso     text        NOT NULL,
    recurso_id  uuid        NOT NULL,
    request_id  text,
    ocurrido_en timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX auditoria_recurso ON auditoria (recurso, recurso_id, ocurrido_en DESC);

-- +goose Down
DROP TABLE auditoria;
DROP TABLE respuestas;
DROP TABLE versiones_formulario;
DROP TABLE formularios;
