// Package memoria implementa repositorios en memoria con datos simulados
// genéricos para el Hito 1 (sin nombres de productos reales).
package memoria

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/service"
)

// Identificadores deterministas de los datos simulados genéricos del Hito 1
// (encuesta, checklist y consentimiento).
const (
	IDFormularioEncuesta       = "10000000-0000-4000-8000-000000000001"
	IDVersionEncuesta          = "20000000-0000-4000-8000-000000000001"
	IDRespuestaEncuesta        = "30000000-0000-4000-8000-000000000001"
	IDFormularioChecklist      = "10000000-0000-4000-8000-000000000002"
	IDVersionChecklist         = "20000000-0000-4000-8000-000000000002"
	IDRespuestaChecklist       = "30000000-0000-4000-8000-000000000002"
	IDFormularioConsentimiento = "10000000-0000-4000-8000-000000000003"
	IDVersionConsentimiento    = "20000000-0000-4000-8000-000000000003"
	IDRespuestaConsentimiento  = "30000000-0000-4000-8000-000000000003"
)

// RepositorioRespuestas almacena respuestas en memoria de forma segura para concurrencia.
type RepositorioRespuestas struct {
	mu    sync.RWMutex
	datos map[string]service.Respuesta
}

// NuevoRepositorioRespuestas crea el repositorio en memoria precargado con ejemplos
// genéricos de encuesta, checklist y consentimiento.
func NuevoRepositorioRespuestas() *RepositorioRespuestas {
	tEncuesta := time.Date(2026, 11, 20, 14, 30, 0, 0, time.UTC)
	tEnvioEncuesta := time.Date(2026, 11, 20, 14, 35, 0, 0, time.UTC)

	tChecklist := time.Date(2026, 11, 21, 10, 15, 0, 0, time.UTC)

	tConsentimiento := time.Date(2026, 11, 22, 16, 0, 0, 0, time.UTC)
	tEnvioConsentimiento := time.Date(2026, 11, 22, 16, 5, 0, 0, time.UTC)

	iniciales := []service.Respuesta{
		{
			ID:            IDRespuestaEncuesta,
			FormularioID:  IDFormularioEncuesta,
			VersionID:     IDVersionEncuesta,
			Estado:        service.EstadoEnviada,
			SujetoTipo:    "atencion",
			SujetoID:      "AT-2026-001",
			Datos:         map[string]any{"calificacion": 5, "recomendaria": true, "comentario": "Atención clara y oportuna."},
			HashContenido: "sha256:6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
			CreadoPor:     service.ActorSimuladoDefecto,
			CreadoEn:      tEncuesta,
			ActualizadoEn: tEnvioEncuesta,
			EnviadaEn:     &tEnvioEncuesta,
		},
		{
			ID:            IDRespuestaChecklist,
			FormularioID:  IDFormularioChecklist,
			VersionID:     IDVersionChecklist,
			Estado:        service.EstadoBorrador,
			SujetoTipo:    "equipo",
			SujetoID:      "EQ-104",
			Datos:         map[string]any{"verificacion_visual": true, "limpieza_completa": true, "observaciones": "Pendiente prueba de encendido."},
			CreadoPor:     service.ActorSimuladoDefecto,
			CreadoEn:      tChecklist,
			ActualizadoEn: tChecklist,
		},
		{
			ID:            IDRespuestaConsentimiento,
			FormularioID:  IDFormularioConsentimiento,
			VersionID:     IDVersionConsentimiento,
			Estado:        service.EstadoEnviada,
			SujetoTipo:    "participante",
			SujetoID:      "PA-0042",
			Datos:         map[string]any{"acepta_terminos": true, "autoriza_contacto": false, "fecha_aceptacion": "2026-11-22"},
			HashContenido: "sha256:d4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab35",
			CreadoPor:     service.ActorSimuladoDefecto,
			CreadoEn:      tConsentimiento,
			ActualizadoEn: tEnvioConsentimiento,
			EnviadaEn:     &tEnvioConsentimiento,
		},
	}

	mapa := make(map[string]service.Respuesta, len(iniciales))
	for _, r := range iniciales {
		mapa[r.ID] = clonarRespuesta(r)
	}

	return &RepositorioRespuestas{datos: mapa}
}

func (r *RepositorioRespuestas) Crear(_ context.Context, resp service.Respuesta) (service.Respuesta, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	copia := clonarRespuesta(resp)
	r.datos[copia.ID] = copia
	return clonarRespuesta(copia), nil
}

func (r *RepositorioRespuestas) ObtenerPorID(_ context.Context, id string) (service.Respuesta, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resp, existe := r.datos[id]
	if !existe {
		return service.Respuesta{}, service.ErrNoEncontrada
	}
	return clonarRespuesta(resp), nil
}

func (r *RepositorioRespuestas) Actualizar(_ context.Context, resp service.Respuesta) (service.Respuesta, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.datos[resp.ID]; !existe {
		return service.Respuesta{}, service.ErrNoEncontrada
	}
	copia := clonarRespuesta(resp)
	r.datos[copia.ID] = copia
	return clonarRespuesta(copia), nil
}

func (r *RepositorioRespuestas) Listar(_ context.Context, filtro service.FiltroRespuestas) ([]service.Respuesta, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	coincidentes := make([]service.Respuesta, 0, len(r.datos))
	for _, item := range r.datos {
		if filtro.FormularioID != "" && item.FormularioID != filtro.FormularioID {
			continue
		}
		if filtro.VersionID != "" && item.VersionID != filtro.VersionID {
			continue
		}
		if filtro.Estado != "" && item.Estado != filtro.Estado {
			continue
		}
		if filtro.SujetoTipo != "" && item.SujetoTipo != filtro.SujetoTipo {
			continue
		}
		if filtro.SujetoID != "" && item.SujetoID != filtro.SujetoID {
			continue
		}
		coincidentes = append(coincidentes, clonarRespuesta(item))
	}

	sort.Slice(coincidentes, func(i, j int) bool {
		if !coincidentes[i].CreadoEn.Equal(coincidentes[j].CreadoEn) {
			return coincidentes[i].CreadoEn.After(coincidentes[j].CreadoEn)
		}
		return coincidentes[i].ID < coincidentes[j].ID
	})

	total := len(coincidentes)
	if filtro.Desplazamiento >= total {
		return []service.Respuesta{}, total, nil
	}

	fin := filtro.Desplazamiento + filtro.Limite
	if fin > total {
		fin = total
	}

	return coincidentes[filtro.Desplazamiento:fin], total, nil
}

func clonarRespuesta(src service.Respuesta) service.Respuesta {
	dst := src
	dst.Datos = clonarMapa(src.Datos)
	if src.EnviadaEn != nil {
		copiaTiempo := *src.EnviadaEn
		dst.EnviadaEn = &copiaTiempo
	}
	return dst
}

func clonarMapa(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = clonarValor(v)
	}
	return dst
}

func clonarValor(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return clonarMapa(t)
	case []any:
		copia := make([]any, len(t))
		for i, elem := range t {
			copia[i] = clonarValor(elem)
		}
		return copia
	default:
		return v
	}
}
