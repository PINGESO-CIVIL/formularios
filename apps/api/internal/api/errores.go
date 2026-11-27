package api

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/danielgtaylor/huma/v2"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/service"
)

const tipoContenidoProblema = "application/problem+json"

// ErrorCampo describe un error de validación de entrada o de esquema de formulario
// asociado a un campo específico mediante un puntero JSON (RFC 6901).
type ErrorCampo struct {
	Location string `json:"location" example:"/datos/calificacion" doc:"Puntero JSON (RFC 6901) del campo con error"`
	Message  string `json:"message" example:"Debe ser un número entero entre 1 y 5." doc:"Mensaje de validación en español"`
	Value    any    `json:"value,omitempty" doc:"Valor recibido en el campo cuando no es sensible"`
}

// ErrorValidacion representa una respuesta RFC 9457 (HTTP 422) con errores por campo.
// Comparte la misma estructura JSON que huma.ErrorModel para mantener consistencia
// entre las validaciones automáticas del contrato y las de esquema de formulario.
type ErrorValidacion struct {
	Type     string       `json:"type,omitempty" format:"uri" example:"about:blank" doc:"URI de referencia del tipo de problema"`
	Title    string       `json:"title" example:"Unprocessable Entity" doc:"Resumen breve del problema"`
	Status   int          `json:"status" example:"422" doc:"Código de estado HTTP"`
	Detail   string       `json:"detail" example:"Los datos de la respuesta no cumplen el esquema del formulario." doc:"Explicación legible en español de esta ocurrencia"`
	Instance string       `json:"instance,omitempty" format:"uri" doc:"URI que identifica esta ocurrencia específica"`
	Errors   []ErrorCampo `json:"errors" doc:"Listado de errores de validación por campo con puntero JSON y mensaje en español"`
}

// documentar422 reemplaza la respuesta 422 generada automáticamente por Huma
// para incluir el esquema ErrorValidacion y los ejemplos por campo visibles en /api/docs.
func documentar422(api huma.API, metodo, ruta, descripcion string, ejemplos map[string]*huma.Example) {
	item := api.OpenAPI().Paths[ruta]
	if item == nil {
		return
	}
	var op *huma.Operation
	switch metodo {
	case http.MethodGet:
		op = item.Get
	case http.MethodPost:
		op = item.Post
	case http.MethodPut:
		op = item.Put
	case http.MethodPatch:
		op = item.Patch
	case http.MethodDelete:
		op = item.Delete
	}
	if op == nil {
		return
	}
	if op.Responses == nil {
		op.Responses = map[string]*huma.Response{}
	}
	op.Responses["422"] = respuesta422Validacion(api, descripcion, ejemplos)
}

// respuesta422Validacion construye la documentación OpenAPI para el código HTTP 422
// con el esquema RFC 9457 y ejemplos de errores por campo visibles en /api/docs.
func respuesta422Validacion(api huma.API, descripcion string, ejemplos map[string]*huma.Example) *huma.Response {
	esquema := api.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[ErrorValidacion](), true, "ErrorValidacion")
	contenido := map[string]*huma.MediaType{
		tipoContenidoProblema: {
			Schema:   esquema,
			Examples: ejemplos,
		},
	}
	return &huma.Response{
		Description: descripcion,
		Content:     contenido,
	}
}

func ejemplosValidacionEnviar() map[string]*huma.Example {
	encuestaInvalida := ErrorValidacion{
		Type:   "about:blank",
		Title:  http.StatusText(http.StatusUnprocessableEntity),
		Status: http.StatusUnprocessableEntity,
		Detail: "Los datos de la respuesta no cumplen el esquema del formulario.",
		Errors: []ErrorCampo{
			{
				Location: "/datos/calificacion",
				Message:  "Debe ser un número entero entre 1 y 5.",
				Value:    7,
			},
			{
				Location: "/datos/correo_contacto",
				Message:  "Debe tener un formato de correo electrónico válido.",
				Value:    "correo-invalido",
			},
		},
	}
	consentimientoNoAceptado := ErrorValidacion{
		Type:   "about:blank",
		Title:  http.StatusText(http.StatusUnprocessableEntity),
		Status: http.StatusUnprocessableEntity,
		Detail: "Los datos de la respuesta no cumplen el esquema del formulario.",
		Errors: []ErrorCampo{
			{
				Location: "/datos/acepta_terminos",
				Message:  "Debe aceptar los términos para enviar el consentimiento.",
				Value:    false,
			},
		},
	}
	return map[string]*huma.Example{
		"encuesta_invalida": {
			Summary:     "Encuesta con campos inválidos",
			Description: "Ejemplo de error 422 al enviar una encuesta con calificación fuera de rango y correo mal formado.",
			Value:       encuestaInvalida,
		},
		"consentimiento_no_aceptado": {
			Summary:     "Consentimiento sin aceptación obligatoria",
			Description: "Ejemplo de error 422 al intentar enviar un consentimiento sin aceptar los términos.",
			Value:       consentimientoNoAceptado,
		},
	}
}

func ejemplosValidacionIniciar() map[string]*huma.Example {
	valor := ErrorValidacion{
		Type:   "about:blank",
		Title:  http.StatusText(http.StatusUnprocessableEntity),
		Status: http.StatusUnprocessableEntity,
		Detail: "La solicitud contiene campos inválidos.",
		Errors: []ErrorCampo{
			{
				Location: "/formulario_id",
				Message:  "Debe ser un identificador UUID válido.",
				Value:    "id-invalido",
			},
			{
				Location: "/sujeto_tipo",
				Message:  "Solo se permiten letras minúsculas, dígitos, guiones y guiones bajos (máximo 60 caracteres).",
				Value:    "Tipo Inválido!",
			},
		},
	}
	return map[string]*huma.Example{
		"identificadores_invalidos": {
			Summary:     "Campos de entrada inválidos al iniciar respuesta",
			Description: "Ejemplo de error 422 cuando el formulario o tipo de sujeto no cumplen las validaciones del contrato.",
			Value:       valor,
		},
	}
}

func ejemplosValidacionBorrador() map[string]*huma.Example {
	valor := ErrorValidacion{
		Type:   "about:blank",
		Title:  http.StatusText(http.StatusUnprocessableEntity),
		Status: http.StatusUnprocessableEntity,
		Detail: "La solicitud contiene campos inválidos.",
		Errors: []ErrorCampo{
			{
				Location: "/id",
				Message:  "Debe ser un identificador UUID válido.",
				Value:    "no-es-uuid",
			},
			{
				Location: "/datos",
				Message:  "El objeto de datos de la respuesta es obligatorio.",
			},
		},
	}
	return map[string]*huma.Example{
		"borrador_invalido": {
			Summary:     "Identificador o cuerpo inválido al guardar borrador",
			Description: "Ejemplo de error 422 cuando el parámetro id no es UUID o falta el objeto datos.",
			Value:       valor,
		},
	}
}

func ejemplosValidacionParametros() map[string]*huma.Example {
	valor := ErrorValidacion{
		Type:   "about:blank",
		Title:  http.StatusText(http.StatusUnprocessableEntity),
		Status: http.StatusUnprocessableEntity,
		Detail: "La solicitud contiene parámetros inválidos.",
		Errors: []ErrorCampo{
			{
				Location: "/estado",
				Message:  "Debe ser uno de los valores permitidos: borrador o enviada.",
				Value:    "pendiente",
			},
		},
	}
	return map[string]*huma.Example{
		"parametro_invalido": {
			Summary:     "Parámetro de consulta o ruta inválido",
			Description: "Ejemplo de error 422 cuando un filtro o identificador no cumple el formato esperado.",
			Value:       valor,
		},
	}
}

// traducirErrorServicio convierte errores del dominio/service en respuestas RFC 9457
// sin exponer trazas internas ni detalles de infraestructura (MF-41).
func traducirErrorServicio(err error) error {
	var errVal *service.ErrorValidacion
	if errors.As(err, &errVal) {
		detalles := make([]error, 0, len(errVal.Errores))
		for _, campo := range errVal.Errores {
			detalles = append(detalles, &huma.ErrorDetail{
				Location: campo.Puntero,
				Message:  campo.Mensaje,
				Value:    campo.Valor,
			})
		}
		return huma.Error422UnprocessableEntity(errVal.Detalle, detalles...)
	}

	switch {
	case errors.Is(err, service.ErrNoEncontrada):
		return huma.Error404NotFound("La respuesta solicitada no existe.")
	case errors.Is(err, service.ErrConflictoEstado):
		return huma.Error409Conflict("La respuesta ya fue enviada y no puede modificarse.")
	case errors.Is(err, service.ErrVersionNoPublicada):
		return huma.Error409Conflict("Solo se puede responder sobre una versión publicada de un formulario activo.")
	case errors.Is(err, service.ErrSinPermiso):
		return huma.Error403Forbidden("No tiene permiso para acceder a esta respuesta.")
	default:
		return huma.Error500InternalServerError("Ocurrió un error interno al procesar la solicitud.")
	}
}
