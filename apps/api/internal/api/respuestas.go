package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/service"
)

const (
	extensionPermiso        = "x-permiso"
	permisoPublico          = "publico"
	permisoRespuestasOperar = "respuestas:operar"
)

// RespuestaDTO representa una respuesta de formulario en el contrato HTTP.
type RespuestaDTO struct {
	ID            string         `json:"id" format:"uuid" example:"30000000-0000-4000-8000-000000000001" doc:"Identificador único de la respuesta"`
	FormularioID  string         `json:"formulario_id" format:"uuid" example:"10000000-0000-4000-8000-000000000001" doc:"Identificador del formulario al que pertenece la respuesta"`
	VersionID     string         `json:"version_id" format:"uuid" example:"20000000-0000-4000-8000-000000000001" doc:"Identificador de la versión publicada sobre la que se responde"`
	Estado        string         `json:"estado" enum:"borrador,enviada" example:"borrador" doc:"Estado actual de la respuesta"`
	SujetoTipo    string         `json:"sujeto_tipo,omitempty" maxLength:"60" example:"atencion" doc:"Tipo opaco de entidad externa asociada, si aplica"`
	SujetoID      string         `json:"sujeto_id,omitempty" maxLength:"120" example:"AT-2026-001" doc:"Identificador opaco de la entidad externa asociada, si aplica"`
	Datos         map[string]any `json:"datos" doc:"Valores respondidos por el usuario según el esquema del formulario"`
	HashContenido string         `json:"hash_contenido,omitempty" example:"sha256:6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b" doc:"Resumen SHA-256 canónico fijado al enviar la respuesta"`
	CreadoPor     string         `json:"creado_por" format:"uuid" example:"99999999-9999-4999-8999-999999999999" doc:"Identificador UUID (claim sub) del usuario creador"`
	CreadoEn      time.Time      `json:"creado_en" example:"2026-11-20T14:30:00Z" doc:"Fecha y hora de creación en formato RFC 3339"`
	ActualizadoEn time.Time      `json:"actualizado_en" example:"2026-11-20T14:35:00Z" doc:"Fecha y hora de última actualización en formato RFC 3339"`
	EnviadaEn     *time.Time     `json:"enviada_en,omitempty" doc:"Fecha y hora de envío definitivo; presente solo cuando estado es enviada"`
}

// IniciarRespuestaInput define la entrada para POST /api/respuestas (RF-FOR-07).
type IniciarRespuestaInput struct {
	Body struct {
		FormularioID string         `json:"formulario_id" format:"uuid" example:"10000000-0000-4000-8000-000000000002" doc:"UUID del formulario activo"`
		VersionID    string         `json:"version_id" format:"uuid" example:"20000000-0000-4000-8000-000000000002" doc:"UUID de la versión publicada sobre la que se responderá"`
		SujetoTipo   string         `json:"sujeto_tipo,omitempty" minLength:"1" maxLength:"60" pattern:"^[a-z0-9_-]+$" example:"equipo" doc:"Tipo opaco de entidad externa asociada (opcional)"`
		SujetoID     string         `json:"sujeto_id,omitempty" minLength:"1" maxLength:"120" example:"EQ-205" doc:"Identificador opaco de la entidad externa asociada (opcional)"`
		Datos        map[string]any `json:"datos,omitempty" doc:"Valores iniciales opcionales para el borrador (ej. encuesta, checklist o consentimiento)"`
	}
}

// GuardarBorradorInput define la entrada para PUT /api/respuestas/{id} (RF-FOR-08).
type GuardarBorradorInput struct {
	ID   string `path:"id" format:"uuid" example:"30000000-0000-4000-8000-000000000002" doc:"UUID de la respuesta en estado borrador"`
	Body struct {
		Datos map[string]any `json:"datos" doc:"Valores actualizados del borrador de respuesta"`
	}
}

// EnviarRespuestaInput define la entrada para POST /api/respuestas/{id}/enviar (RF-FOR-09).
type EnviarRespuestaInput struct {
	ID   string `path:"id" format:"uuid" example:"30000000-0000-4000-8000-000000000002" doc:"UUID de la respuesta en borrador a enviar"`
	Body struct {
		Datos map[string]any `json:"datos,omitempty" doc:"Valores finales opcionales a validar antes del envío; si se omite, se validan los datos guardados en el borrador"`
	}
}

// ObtenerRespuestaInput define la entrada para GET /api/respuestas/{id} (RF-FOR-10).
type ObtenerRespuestaInput struct {
	ID string `path:"id" format:"uuid" example:"30000000-0000-4000-8000-000000000001" doc:"UUID de la respuesta a consultar"`
}

// ListarRespuestasInput define los filtros y paginación para GET /api/respuestas (RF-FOR-10).
type ListarRespuestasInput struct {
	FormularioID   string `query:"formulario_id" format:"uuid" doc:"Filtra por UUID del formulario"`
	VersionID      string `query:"version_id" format:"uuid" doc:"Filtra por UUID de la versión publicada"`
	Estado         string `query:"estado" enum:"borrador,enviada" doc:"Filtra por estado de la respuesta"`
	SujetoTipo     string `query:"sujeto_tipo" maxLength:"60" pattern:"^[a-z0-9_-]+$" doc:"Filtra por tipo opaco de entidad externa"`
	SujetoID       string `query:"sujeto_id" maxLength:"120" doc:"Filtra por identificador opaco de entidad externa"`
	Limite         int    `query:"limite" minimum:"1" maximum:"100" default:"20" doc:"Cantidad máxima de resultados por página"`
	Desplazamiento int    `query:"desplazamiento" minimum:"0" maximum:"10000" default:"0" doc:"Cantidad de registros a omitir"`
}

// RespuestaOutput envuelve una única respuesta en el cuerpo JSON.
type RespuestaOutput struct {
	Body RespuestaDTO
}

// ListarRespuestasOutput envuelve el listado paginado de respuestas.
type ListarRespuestasOutput struct {
	Body struct {
		Elementos      []RespuestaDTO `json:"elementos" doc:"Respuestas que cumplen los criterios de búsqueda"`
		Total          int            `json:"total" example:"3" doc:"Cantidad total de respuestas coincidentes"`
		Limite         int            `json:"limite" example:"20" doc:"Límite aplicado en esta página"`
		Desplazamiento int            `json:"desplazamiento" example:"0" doc:"Desplazamiento aplicado en esta página"`
	}
}

// conPermiso adjunta la extensión x-permiso exigida por el DAS §6.
func conPermiso(op huma.Operation, permiso string) huma.Operation {
	if op.Extensions == nil {
		op.Extensions = map[string]any{}
	}
	op.Extensions[extensionPermiso] = permiso
	return op
}

// registrarRespuestas registra los cinco endpoints del recurso /api/respuestas.
// Solo traduce HTTP <-> service: nunca importa ni invoca internal/store.
func registrarRespuestas(api huma.API, svc service.ServicioRespuestas) {
	erroresMutacion := []int{
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusConflict,
	}
	erroresConsulta := []int{
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
	}
	erroresListado := []int{
		http.StatusUnauthorized,
		http.StatusForbidden,
	}

	huma.Register(api, conPermiso(huma.Operation{
		OperationID:   "iniciarRespuesta",
		Method:        http.MethodPost,
		Path:          "/api/respuestas",
		Summary:       "Inicia una respuesta en estado borrador",
		Description:   "Crea una respuesta en estado borrador asociada a una versión publicada de un formulario activo (RF-FOR-07).",
		Tags:          []string{"Respuestas"},
		DefaultStatus: http.StatusCreated,
		Errors:        erroresMutacion,
	}, permisoRespuestasOperar), func(ctx context.Context, input *IniciarRespuestaInput) (*RespuestaOutput, error) {
		if svc == nil {
			return nil, huma.Error503ServiceUnavailable("El servicio de respuestas no está disponible.")
		}
		resp, err := svc.Iniciar(ctx, service.IniciarRespuestaEntrada{
			FormularioID: input.Body.FormularioID,
			VersionID:    input.Body.VersionID,
			SujetoTipo:   input.Body.SujetoTipo,
			SujetoID:     input.Body.SujetoID,
			Datos:        input.Body.Datos,
		})
		if err != nil {
			return nil, traducirErrorServicio(err)
		}
		return &RespuestaOutput{Body: aRespuestaDTO(resp)}, nil
	})
	documentar422(api, http.MethodPost, "/api/respuestas", "Error de validación en los campos de la solicitud", ejemplosValidacionIniciar())

	huma.Register(api, conPermiso(huma.Operation{
		OperationID: "guardarBorradorRespuesta",
		Method:      http.MethodPut,
		Path:        "/api/respuestas/{id}",
		Summary:     "Guarda el avance parcial de una respuesta en borrador",
		Description: "Actualiza los datos de una respuesta mientras se encuentre en estado borrador (RF-FOR-08). Responde 409 si ya fue enviada.",
		Tags:        []string{"Respuestas"},
		Errors:      erroresMutacion,
	}, permisoRespuestasOperar), func(ctx context.Context, input *GuardarBorradorInput) (*RespuestaOutput, error) {
		if svc == nil {
			return nil, huma.Error503ServiceUnavailable("El servicio de respuestas no está disponible.")
		}
		resp, err := svc.GuardarBorrador(ctx, service.GuardarBorradorEntrada{
			ID:    input.ID,
			Datos: input.Body.Datos,
		})
		if err != nil {
			return nil, traducirErrorServicio(err)
		}
		return &RespuestaOutput{Body: aRespuestaDTO(resp)}, nil
	})
	documentar422(api, http.MethodPut, "/api/respuestas/{id}", "Error de validación en los campos del borrador", ejemplosValidacionBorrador())

	huma.Register(api, conPermiso(huma.Operation{
		OperationID: "enviarRespuesta",
		Method:      http.MethodPost,
		Path:        "/api/respuestas/{id}/enviar",
		Summary:     "Valida y envía definitivamente una respuesta",
		Description: "Valida los datos contra el esquema de la versión publicada, calcula hash_contenido y cambia el estado a enviada de forma inmutable (RF-FOR-09). Si hay errores de validación devuelve 422 con errores por campo (puntero JSON y mensaje en español).",
		Tags:        []string{"Respuestas"},
		Errors:      erroresMutacion,
	}, permisoRespuestasOperar), func(ctx context.Context, input *EnviarRespuestaInput) (*RespuestaOutput, error) {
		if svc == nil {
			return nil, huma.Error503ServiceUnavailable("El servicio de respuestas no está disponible.")
		}
		resp, err := svc.Enviar(ctx, service.EnviarRespuestaEntrada{
			ID:    input.ID,
			Datos: input.Body.Datos,
		})
		if err != nil {
			return nil, traducirErrorServicio(err)
		}
		return &RespuestaOutput{Body: aRespuestaDTO(resp)}, nil
	})
	documentar422(api, http.MethodPost, "/api/respuestas/{id}/enviar", "Los datos de la respuesta no cumplen el esquema del formulario", ejemplosValidacionEnviar())

	huma.Register(api, conPermiso(huma.Operation{
		OperationID: "obtenerRespuesta",
		Method:      http.MethodGet,
		Path:        "/api/respuestas/{id}",
		Summary:     "Obtiene el detalle de una respuesta por UUID",
		Description: "Consulta una respuesta por su identificador UUID (RF-FOR-10). Con respuestas:operar permite ver las propias; con respuestas:administrar permite ver respuestas de cualquier usuario.",
		Tags:        []string{"Respuestas"},
		Errors:      erroresConsulta,
	}, permisoRespuestasOperar), func(ctx context.Context, input *ObtenerRespuestaInput) (*RespuestaOutput, error) {
		if svc == nil {
			return nil, huma.Error503ServiceUnavailable("El servicio de respuestas no está disponible.")
		}
		resp, err := svc.Obtener(ctx, input.ID)
		if err != nil {
			return nil, traducirErrorServicio(err)
		}
		return &RespuestaOutput{Body: aRespuestaDTO(resp)}, nil
	})
	documentar422(api, http.MethodGet, "/api/respuestas/{id}", "Identificador de respuesta inválido", ejemplosValidacionParametros())

	huma.Register(api, conPermiso(huma.Operation{
		OperationID: "listarRespuestas",
		Method:      http.MethodGet,
		Path:        "/api/respuestas",
		Summary:     "Lista respuestas con filtros y paginación",
		Description: "Devuelve un listado paginado de respuestas filtrable por formulario, versión, estado y sujeto externo (RF-FOR-10). Con respuestas:operar filtra por el usuario autenticado; con respuestas:administrar lista todas.",
		Tags:        []string{"Respuestas"},
		Errors:      erroresListado,
	}, permisoRespuestasOperar), func(ctx context.Context, input *ListarRespuestasInput) (*ListarRespuestasOutput, error) {
		if svc == nil {
			return nil, huma.Error503ServiceUnavailable("El servicio de respuestas no está disponible.")
		}
		pagina, err := svc.Listar(ctx, service.FiltroRespuestas{
			FormularioID:   input.FormularioID,
			VersionID:      input.VersionID,
			Estado:         service.EstadoRespuesta(input.Estado),
			SujetoTipo:     input.SujetoTipo,
			SujetoID:       input.SujetoID,
			Limite:         input.Limite,
			Desplazamiento: input.Desplazamiento,
		})
		if err != nil {
			return nil, traducirErrorServicio(err)
		}

		salida := &ListarRespuestasOutput{}
		salida.Body.Elementos = make([]RespuestaDTO, 0, len(pagina.Elementos))
		for _, elem := range pagina.Elementos {
			salida.Body.Elementos = append(salida.Body.Elementos, aRespuestaDTO(elem))
		}
		salida.Body.Total = pagina.Total
		salida.Body.Limite = pagina.Limite
		salida.Body.Desplazamiento = pagina.Desplazamiento
		return salida, nil
	})
	documentar422(api, http.MethodGet, "/api/respuestas", "Parámetros de consulta inválidos", ejemplosValidacionParametros())
}

func aRespuestaDTO(r service.Respuesta) RespuestaDTO {
	datos := r.Datos
	if datos == nil {
		datos = map[string]any{}
	}
	return RespuestaDTO{
		ID:            r.ID,
		FormularioID:  r.FormularioID,
		VersionID:     r.VersionID,
		Estado:        string(r.Estado),
		SujetoTipo:    r.SujetoTipo,
		SujetoID:      r.SujetoID,
		Datos:         datos,
		HashContenido: r.HashContenido,
		CreadoPor:     r.CreadoPor,
		CreadoEn:      r.CreadoEn,
		ActualizadoEn: r.ActualizadoEn,
		EnviadaEn:     r.EnviadaEn,
	}
}
