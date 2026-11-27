package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/service"
)

const (
	uuidFormularioEncuesta  = "10000000-0000-4000-8000-000000000001"
	uuidVersionEncuesta     = "20000000-0000-4000-8000-000000000001"
	uuidRespuestaEncuesta   = "30000000-0000-4000-8000-000000000001"
	uuidFormularioChecklist = "10000000-0000-4000-8000-000000000002"
	uuidVersionChecklist    = "20000000-0000-4000-8000-000000000002"
	uuidRespuestaChecklist  = "30000000-0000-4000-8000-000000000002"
)

// repoPrueba implementa service.RepositorioRespuestas para aislar las pruebas
// de internal/api sin importar internal/store.
type repoPrueba struct {
	items map[string]service.Respuesta
}

func nuevoRepoPrueba() *repoPrueba {
	t := time.Date(2026, 11, 20, 14, 30, 0, 0, time.UTC)
	tEnvio := time.Date(2026, 11, 20, 14, 35, 0, 0, time.UTC)
	return &repoPrueba{
		items: map[string]service.Respuesta{
			uuidRespuestaEncuesta: {
				ID:            uuidRespuestaEncuesta,
				FormularioID:  uuidFormularioEncuesta,
				VersionID:     uuidVersionEncuesta,
				Estado:        service.EstadoEnviada,
				SujetoTipo:    "atencion",
				SujetoID:      "AT-2026-001",
				Datos:         map[string]any{"calificacion": 5, "recomendaria": true},
				HashContenido: "sha256:6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				CreadoPor:     service.ActorSimuladoDefecto,
				CreadoEn:      t,
				ActualizadoEn: tEnvio,
				EnviadaEn:     &tEnvio,
			},
			uuidRespuestaChecklist: {
				ID:            uuidRespuestaChecklist,
				FormularioID:  uuidFormularioChecklist,
				VersionID:     uuidVersionChecklist,
				Estado:        service.EstadoBorrador,
				SujetoTipo:    "equipo",
				SujetoID:      "EQ-104",
				Datos:         map[string]any{"verificacion_visual": true},
				CreadoPor:     service.ActorSimuladoDefecto,
				CreadoEn:      t.Add(time.Hour),
				ActualizadoEn: t.Add(time.Hour),
			},
		},
	}
}

func (r *repoPrueba) Crear(_ context.Context, resp service.Respuesta) (service.Respuesta, error) {
	r.items[resp.ID] = resp
	return resp, nil
}

func (r *repoPrueba) ObtenerPorID(_ context.Context, id string) (service.Respuesta, error) {
	resp, ok := r.items[id]
	if !ok {
		return service.Respuesta{}, service.ErrNoEncontrada
	}
	return resp, nil
}

func (r *repoPrueba) Actualizar(_ context.Context, resp service.Respuesta) (service.Respuesta, error) {
	if _, ok := r.items[resp.ID]; !ok {
		return service.Respuesta{}, service.ErrNoEncontrada
	}
	r.items[resp.ID] = resp
	return resp, nil
}

func (r *repoPrueba) Listar(_ context.Context, filtro service.FiltroRespuestas) ([]service.Respuesta, int, error) {
	var lista []service.Respuesta
	for _, item := range r.items {
		if filtro.Estado != "" && item.Estado != filtro.Estado {
			continue
		}
		if filtro.FormularioID != "" && item.FormularioID != filtro.FormularioID {
			continue
		}
		lista = append(lista, item)
	}
	return lista, len(lista), nil
}

func nuevoRouterRespuestas(publicarDocs bool) (http.Handler, huma.API) {
	svc := service.NuevoServicioRespuestasSimulado(nuevoRepoPrueba())
	return servidor(RouterDeps{
		Metrics:      prometheus.NewRegistry(),
		PublicarDocs: publicarDocs,
		Respuestas:   svc,
	})
}

func hacerPeticion(h http.Handler, metodo, ruta string, cuerpo any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if cuerpo != nil {
		_ = json.NewEncoder(&buf).Encode(cuerpo)
	}
	req := httptest.NewRequestWithContext(context.Background(), metodo, ruta, &buf)
	if cuerpo != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// MF-13: cada operación publicada en el contrato debe declarar su permiso en x-permiso.
func TestCadaOperacionDeclaraPermiso(t *testing.T) {
	_, humaAPI := nuevoRouterRespuestas(true)
	require.NotEmpty(t, humaAPI.OpenAPI().Paths)

	for ruta, item := range humaAPI.OpenAPI().Paths {
		operaciones := map[string]*huma.Operation{
			http.MethodGet:    item.Get,
			http.MethodPost:   item.Post,
			http.MethodPut:    item.Put,
			http.MethodPatch:  item.Patch,
			http.MethodDelete: item.Delete,
		}
		for metodo, op := range operaciones {
			if op == nil {
				continue
			}
			permiso, ok := op.Extensions[extensionPermiso].(string)
			assert.True(t, ok && permiso != "", "%s %s debe declarar x-permiso", metodo, ruta)
		}
	}
}

// MF-05 y criterio de aceptación #6: /api/docs solo fuera de prod y con ejemplo 422 por campo.
func TestDocsExponeEjemplo422ConErroresPorCampo(t *testing.T) {
	handlerDev, humaAPI := nuevoRouterRespuestas(true)
	recDev := hacerPeticion(handlerDev, http.MethodGet, "/api/docs", nil)
	assert.Equal(t, http.StatusOK, recDev.Code)

	handlerProd, _ := nuevoRouterRespuestas(false)
	recProd := hacerPeticion(handlerProd, http.MethodGet, "/api/docs", nil)
	assert.Equal(t, http.StatusNotFound, recProd.Code)

	opEnviar := humaAPI.OpenAPI().Paths["/api/respuestas/{id}/enviar"].Post
	require.NotNil(t, opEnviar)
	resp422 := opEnviar.Responses["422"]
	require.NotNil(t, resp422)

	media := resp422.Content[tipoContenidoProblema]
	require.NotNil(t, media)
	require.NotNil(t, media.Schema)
	require.NotEmpty(t, media.Examples)

	ejemplo, existe := media.Examples["encuesta_invalida"]
	require.True(t, existe)
	valor, ok := ejemplo.Value.(ErrorValidacion)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, valor.Status)
	require.NotEmpty(t, valor.Errors)
	assert.True(t, strings.HasPrefix(valor.Errors[0].Location, "/datos/"))
	assert.NotEmpty(t, valor.Errors[0].Message)
}

func TestCicloVidaRespuestaSimulada(t *testing.T) {
	h, _ := nuevoRouterRespuestas(true)

	// 1. Iniciar borrador (RF-FOR-07).
	recIniciar := hacerPeticion(h, http.MethodPost, "/api/respuestas", map[string]any{
		"formulario_id": uuidFormularioEncuesta,
		"version_id":    uuidVersionEncuesta,
		"sujeto_tipo":   "atencion",
		"sujeto_id":     "AT-2026-099",
		"datos":         map[string]any{"calificacion": 7},
	})
	require.Equal(t, http.StatusCreated, recIniciar.Code)

	var creada RespuestaDTO
	require.NoError(t, json.Unmarshal(recIniciar.Body.Bytes(), &creada))
	assert.Equal(t, "borrador", creada.Estado)
	assert.NotEmpty(t, creada.ID)

	// 2. Enviar con datos inválidos -> 422 RFC 9457 con puntero JSON y mensaje en español (RN-FOR-06).
	recInvalido := hacerPeticion(h, http.MethodPost, "/api/respuestas/"+creada.ID+"/enviar", map[string]any{})
	require.Equal(t, http.StatusUnprocessableEntity, recInvalido.Code)
	assert.Contains(t, recInvalido.Header().Get("Content-Type"), tipoContenidoProblema)

	var errVal ErrorValidacion
	require.NoError(t, json.Unmarshal(recInvalido.Body.Bytes(), &errVal))
	require.NotEmpty(t, errVal.Errors)
	assert.Equal(t, "/datos/calificacion", errVal.Errors[0].Location)
	assert.Contains(t, errVal.Errors[0].Message, "entre 1 y 5")

	// 3. Guardar borrador corregido (RF-FOR-08).
	recGuardar := hacerPeticion(h, http.MethodPut, "/api/respuestas/"+creada.ID, map[string]any{
		"datos": map[string]any{"calificacion": 5, "recomendaria": true},
	})
	require.Equal(t, http.StatusOK, recGuardar.Code)

	// 4. Enviar definitivamente (RF-FOR-09).
	recEnviar := hacerPeticion(h, http.MethodPost, "/api/respuestas/"+creada.ID+"/enviar", map[string]any{})
	require.Equal(t, http.StatusOK, recEnviar.Code)

	var enviada RespuestaDTO
	require.NoError(t, json.Unmarshal(recEnviar.Body.Bytes(), &enviada))
	assert.Equal(t, "enviada", enviada.Estado)
	assert.NotEmpty(t, enviada.HashContenido)
	assert.NotNil(t, enviada.EnviadaEn)

	// 5. Intentar editar o reenviar una respuesta enviada -> 409 (RN-FOR-04 y RN-FOR-07).
	recEditarEnviada := hacerPeticion(h, http.MethodPut, "/api/respuestas/"+creada.ID, map[string]any{
		"datos": map[string]any{"calificacion": 1},
	})
	assert.Equal(t, http.StatusConflict, recEditarEnviada.Code)

	recReenviar := hacerPeticion(h, http.MethodPost, "/api/respuestas/"+creada.ID+"/enviar", map[string]any{})
	assert.Equal(t, http.StatusConflict, recReenviar.Code)
}

func TestConsultarYListarRespuestas(t *testing.T) {
	h, _ := nuevoRouterRespuestas(true)

	recListar := hacerPeticion(h, http.MethodGet, "/api/respuestas?estado=borrador", nil)
	require.Equal(t, http.StatusOK, recListar.Code)
	assert.Contains(t, recListar.Body.String(), uuidRespuestaChecklist)

	recVer := hacerPeticion(h, http.MethodGet, "/api/respuestas/"+uuidRespuestaEncuesta, nil)
	require.Equal(t, http.StatusOK, recVer.Code)
	assert.Contains(t, recVer.Body.String(), "AT-2026-001")

	recNoExiste := hacerPeticion(h, http.MethodGet, "/api/respuestas/99999999-9999-4999-8999-999999999998", nil)
	assert.Equal(t, http.StatusNotFound, recNoExiste.Code)
	assert.Contains(t, recNoExiste.Header().Get("Content-Type"), tipoContenidoProblema)

	recUUIDInvalido := hacerPeticion(h, http.MethodGet, "/api/respuestas/no-es-uuid", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, recUUIDInvalido.Code)
}
