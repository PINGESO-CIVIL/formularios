package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// EstadoRespuesta representa el estado del ciclo de vida de una respuesta.
type EstadoRespuesta string

const (
	EstadoBorrador EstadoRespuesta = "borrador"
	EstadoEnviada  EstadoRespuesta = "enviada"
)

// ActorSimuladoDefecto es el UUID temporal usado en Hito 1 mientras no se
// integre el middleware JWT del issue #4.
// TODO(integracion): extraer el claim sub del JWT validado por usuarios-roles (issue #4).
const ActorSimuladoDefecto = "99999999-9999-4999-8999-999999999999"

const (
	limitePorDefecto = 20
	limiteMaximo     = 100
)

var (
	ErrNoEncontrada       = errors.New("respuesta no encontrada")
	ErrConflictoEstado    = errors.New("la respuesta ya fue enviada y no puede modificarse")
	ErrVersionNoPublicada = errors.New("solo se puede responder sobre una versión publicada de un formulario activo")
	ErrSinPermiso         = errors.New("no tiene permiso para acceder a esta respuesta")
)

// ErrorCampo describe un error de validación asociado a un campo específico.
type ErrorCampo struct {
	Puntero string
	Mensaje string
	Valor   any
}

// ErrorValidacion agrupa los errores de validación por campo (HTTP 422).
type ErrorValidacion struct {
	Detalle string
	Errores []ErrorCampo
}

func (e *ErrorValidacion) Error() string {
	if e.Detalle != "" {
		return e.Detalle
	}
	return "los datos de la respuesta no son válidos"
}

// Respuesta representa el modelo de negocio de una respuesta de formulario.
type Respuesta struct {
	ID            string
	FormularioID  string
	VersionID     string
	Estado        EstadoRespuesta
	SujetoTipo    string
	SujetoID      string
	Datos         map[string]any
	HashContenido string
	CreadoPor     string
	CreadoEn      time.Time
	ActualizadoEn time.Time
	EnviadaEn     *time.Time
}

// IniciarRespuestaEntrada contiene los datos necesarios para crear un borrador de respuesta.
type IniciarRespuestaEntrada struct {
	FormularioID string
	VersionID    string
	SujetoTipo   string
	SujetoID     string
	Datos        map[string]any
	ActorID      string
}

// GuardarBorradorEntrada contiene los datos para actualizar una respuesta en borrador.
type GuardarBorradorEntrada struct {
	ID      string
	Datos   map[string]any
	ActorID string
}

// EnviarRespuestaEntrada contiene los datos para validar y enviar definitivamente una respuesta.
type EnviarRespuestaEntrada struct {
	ID      string
	Datos   map[string]any
	ActorID string
}

// FiltroRespuestas define filtros y paginación para consultar respuestas.
type FiltroRespuestas struct {
	FormularioID    string
	VersionID       string
	Estado          EstadoRespuesta
	SujetoTipo      string
	SujetoID        string
	Limite          int
	Desplazamiento  int
	ActorID         string
	EsAdministrador bool
}

// PaginaRespuestas agrupa los resultados paginados de un listado de respuestas.
type PaginaRespuestas struct {
	Elementos      []Respuesta
	Total          int
	Limite         int
	Desplazamiento int
}

// RepositorioRespuestas define el puerto de almacenamiento que consume el service.
// En Hito 1 se implementa en memoria (internal/store/memoria); en Hito 2 sobre PostgreSQL.
type RepositorioRespuestas interface {
	Crear(ctx context.Context, r Respuesta) (Respuesta, error)
	ObtenerPorID(ctx context.Context, id string) (Respuesta, error)
	Actualizar(ctx context.Context, r Respuesta) (Respuesta, error)
	Listar(ctx context.Context, filtro FiltroRespuestas) ([]Respuesta, int, error)
}

// ServicioRespuestas define el contrato de negocio consumido por los handlers HTTP.
type ServicioRespuestas interface {
	Iniciar(ctx context.Context, entrada IniciarRespuestaEntrada) (Respuesta, error)
	GuardarBorrador(ctx context.Context, entrada GuardarBorradorEntrada) (Respuesta, error)
	Enviar(ctx context.Context, entrada EnviarRespuestaEntrada) (Respuesta, error)
	Obtener(ctx context.Context, id string) (Respuesta, error)
	Listar(ctx context.Context, filtro FiltroRespuestas) (PaginaRespuestas, error)
}

// RespuestasSimuladas implementa ServicioRespuestas para el Hito 1 sobre un repositorio en memoria.
type RespuestasSimuladas struct {
	repo  RepositorioRespuestas
	ahora func() time.Time
}

// NuevoServicioRespuestasSimulado construye el servicio simulado de respuestas para el Hito 1.
func NuevoServicioRespuestasSimulado(repo RepositorioRespuestas) *RespuestasSimuladas {
	return &RespuestasSimuladas{
		repo:  repo,
		ahora: horaActualUTC,
	}
}

func horaActualUTC() time.Time {
	return time.Now().UTC()
}

func (s *RespuestasSimuladas) Iniciar(ctx context.Context, entrada IniciarRespuestaEntrada) (Respuesta, error) {
	id, err := nuevoUUID()
	if err != nil {
		return Respuesta{}, err
	}

	actor := strings.TrimSpace(entrada.ActorID)
	if actor == "" {
		actor = ActorSimuladoDefecto
	}

	datos := entrada.Datos
	if datos == nil {
		datos = map[string]any{}
	}

	instante := s.ahora()
	nueva := Respuesta{
		ID:            id,
		FormularioID:  entrada.FormularioID,
		VersionID:     entrada.VersionID,
		Estado:        EstadoBorrador,
		SujetoTipo:    entrada.SujetoTipo,
		SujetoID:      entrada.SujetoID,
		Datos:         datos,
		CreadoPor:     actor,
		CreadoEn:      instante,
		ActualizadoEn: instante,
	}

	return s.repo.Crear(ctx, nueva)
}

func (s *RespuestasSimuladas) GuardarBorrador(ctx context.Context, entrada GuardarBorradorEntrada) (Respuesta, error) {
	actual, err := s.repo.ObtenerPorID(ctx, entrada.ID)
	if err != nil {
		return Respuesta{}, err
	}

	// RN-FOR-04: una respuesta enviada es inmutable.
	if actual.Estado == EstadoEnviada {
		return Respuesta{}, ErrConflictoEstado
	}

	if entrada.Datos == nil {
		actual.Datos = map[string]any{}
	} else {
		actual.Datos = entrada.Datos
	}
	actual.ActualizadoEn = s.ahora()

	return s.repo.Actualizar(ctx, actual)
}

func (s *RespuestasSimuladas) Enviar(ctx context.Context, entrada EnviarRespuestaEntrada) (Respuesta, error) {
	actual, err := s.repo.ObtenerPorID(ctx, entrada.ID)
	if err != nil {
		return Respuesta{}, err
	}

	// RN-FOR-04 y RN-FOR-07: una respuesta ya enviada no puede reenviarse.
	if actual.Estado == EstadoEnviada {
		return Respuesta{}, ErrConflictoEstado
	}

	datosAValidar := actual.Datos
	if entrada.Datos != nil {
		datosAValidar = entrada.Datos
	}

	// Validación simulada de Hito 1 (RN-FOR-06). En Hito 2 se reemplaza por validación
	// real contra el JSON Schema 2020-12 de la versión publicada.
	if erroresCampo := validarDatosSimulados(datosAValidar); len(erroresCampo) > 0 {
		return Respuesta{}, &ErrorValidacion{
			Detalle: "Los datos de la respuesta no cumplen el esquema del formulario.",
			Errores: erroresCampo,
		}
	}

	hash, err := calcularHashContenido(datosAValidar)
	if err != nil {
		return Respuesta{}, err
	}

	instante := s.ahora()
	actual.Datos = datosAValidar
	actual.Estado = EstadoEnviada
	actual.HashContenido = hash
	actual.ActualizadoEn = instante
	actual.EnviadaEn = &instante

	return s.repo.Actualizar(ctx, actual)
}

func (s *RespuestasSimuladas) Obtener(ctx context.Context, id string) (Respuesta, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *RespuestasSimuladas) Listar(ctx context.Context, filtro FiltroRespuestas) (PaginaRespuestas, error) {
	if filtro.Limite <= 0 {
		filtro.Limite = limitePorDefecto
	}
	if filtro.Limite > limiteMaximo {
		filtro.Limite = limiteMaximo
	}
	if filtro.Desplazamiento < 0 {
		filtro.Desplazamiento = 0
	}

	elementos, total, err := s.repo.Listar(ctx, filtro)
	if err != nil {
		return PaginaRespuestas{}, err
	}
	if elementos == nil {
		elementos = []Respuesta{}
	}

	return PaginaRespuestas{
		Elementos:      elementos,
		Total:          total,
		Limite:         filtro.Limite,
		Desplazamiento: filtro.Desplazamiento,
	}, nil
}

func validarDatosSimulados(datos map[string]any) []ErrorCampo {
	if len(datos) == 0 {
		return []ErrorCampo{
			{
				Puntero: "/datos",
				Mensaje: "Debe incluir al menos un campo respondido antes de enviar.",
			},
		}
	}

	var erroresCampo []ErrorCampo

	if valor, existe := datos["calificacion"]; existe {
		if !esEnteroEnRango(valor, 1, 5) {
			erroresCampo = append(erroresCampo, ErrorCampo{
				Puntero: "/datos/calificacion",
				Mensaje: "Debe ser un número entero entre 1 y 5.",
				Valor:   valor,
			})
		}
	}

	if valor, existe := datos["acepta_terminos"]; existe {
		aceptado, ok := valor.(bool)
		if !ok || !aceptado {
			erroresCampo = append(erroresCampo, ErrorCampo{
				Puntero: "/datos/acepta_terminos",
				Mensaje: "Debe aceptar los términos para enviar el consentimiento.",
				Valor:   valor,
			})
		}
	}

	if valor, existe := datos["correo_contacto"]; existe {
		correo, ok := valor.(string)
		if !ok || !strings.Contains(correo, "@") || strings.Contains(correo, " ") {
			erroresCampo = append(erroresCampo, ErrorCampo{
				Puntero: "/datos/correo_contacto",
				Mensaje: "Debe tener un formato de correo electrónico válido.",
				Valor:   valor,
			})
		}
	}

	return erroresCampo
}

func esEnteroEnRango(valor any, minimo, maximo int) bool {
	switch v := valor.(type) {
	case int:
		return v >= minimo && v <= maximo
	case int64:
		return v >= int64(minimo) && v <= int64(maximo)
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v {
			return false
		}
		entero := int(v)
		return entero >= minimo && entero <= maximo
	default:
		return false
	}
}

func calcularHashContenido(datos map[string]any) (string, error) {
	crudo, err := json.Marshal(datos)
	if err != nil {
		return "", fmt.Errorf("serializar datos para hash: %w", err)
	}
	suma := sha256.Sum256(crudo)
	return "sha256:" + hex.EncodeToString(suma[:]), nil
}

func nuevoUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generar UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
