package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"time"
)

// cabeceraRequestID es la cabecera de correlación entre servicios (DAS §8).
const cabeceraRequestID = "X-Request-Id"

// requestIDValido limita lo que se acepta desde afuera, para no escribir basura en los logs.
var requestIDValido = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

type claveContexto struct{}

// RequestID devuelve el id de la petición guardado en el contexto, o "" si no hay.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(claveContexto{}).(string)
	return id
}

// conObservabilidad asigna o propaga X-Request-Id y registra cada petición con los campos
// del DAS: request_id, metodo, ruta, estado y duracion_ms. Nunca registra cuerpos,
// cabeceras de autorización ni cookies.
func conObservabilidad(siguiente http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		id := r.Header.Get(cabeceraRequestID)
		if !requestIDValido.MatchString(id) {
			id = nuevoRequestID()
		}
		w.Header().Set(cabeceraRequestID, id)

		grabador := &grabadorEstado{ResponseWriter: w, estado: http.StatusOK}
		siguiente.ServeHTTP(grabador, r.WithContext(context.WithValue(r.Context(), claveContexto{}, id)))

		if logger != nil {
			logger.LogAttrs(r.Context(), slog.LevelInfo, "petición atendida",
				slog.String("request_id", id),
				slog.String("metodo", r.Method),
				slog.String("ruta", r.URL.Path),
				slog.Int("estado", grabador.estado),
				slog.Int64("duracion_ms", time.Since(inicio).Milliseconds()),
			)
		}
	})
}

// nuevoRequestID genera un id aleatorio de 16 bytes en hexadecimal.
func nuevoRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// grabadorEstado recuerda el código de estado que escribió el handler.
type grabadorEstado struct {
	http.ResponseWriter
	estado int
}

func (g *grabadorEstado) WriteHeader(estado int) {
	g.estado = estado
	g.ResponseWriter.WriteHeader(estado)
}
