package memoria_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/service"
	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/store/memoria"
)

func TestRepositorioMemoriaPrecargaDatosGenericosSinNombresDeProductos(t *testing.T) {
	ctx := context.Background()
	repo := memoria.NuevoRepositorioRespuestas()
	svc := service.NuevoServicioRespuestasSimulado(repo)

	pagina, err := svc.Listar(ctx, service.FiltroRespuestas{})
	require.NoError(t, err)
	assert.Equal(t, 3, pagina.Total)
	assert.Len(t, pagina.Elementos, 3)

	crudo, err := json.Marshal(pagina.Elementos)
	require.NoError(t, err)
	contenido := strings.ToLower(string(crudo))

	prohibidos := []string{"inverfit", "mad360", "gatica", "skinlaser", "to usach"}
	for _, nombre := range prohibidos {
		assert.NotContains(t, contenido, nombre, "los datos simulados no deben mencionar productos reales")
	}

	assert.Contains(t, contenido, "calificacion")
	assert.Contains(t, contenido, "verificacion_visual")
	assert.Contains(t, contenido, "acepta_terminos")
}

func TestServicioRespuestasSimuladoValidaYProtegeInmutabilidad(t *testing.T) {
	ctx := context.Background()
	svc := service.NuevoServicioRespuestasSimulado(memoria.NuevoRepositorioRespuestas())

	// Enviar consentimiento con acepta_terminos=false devuelve ErrorValidacion por campo (RN-FOR-06).
	creada, err := svc.Iniciar(ctx, service.IniciarRespuestaEntrada{
		FormularioID: memoria.IDFormularioConsentimiento,
		VersionID:    memoria.IDVersionConsentimiento,
		SujetoTipo:   "participante",
		SujetoID:     "PA-0099",
		Datos:        map[string]any{"acepta_terminos": false},
	})
	require.NoError(t, err)

	_, err = svc.Enviar(ctx, service.EnviarRespuestaEntrada{ID: creada.ID})
	require.Error(t, err)

	var errVal *service.ErrorValidacion
	require.True(t, errors.As(err, &errVal))
	require.Len(t, errVal.Errores, 1)
	assert.Equal(t, "/datos/acepta_terminos", errVal.Errores[0].Puntero)

	// Una respuesta ya enviada no puede volver a editarse ni enviarse (RN-FOR-04).
	_, err = svc.GuardarBorrador(ctx, service.GuardarBorradorEntrada{
		ID:    memoria.IDRespuestaEncuesta,
		Datos: map[string]any{"calificacion": 1},
	})
	assert.ErrorIs(t, err, service.ErrConflictoEstado)

	_, err = svc.Enviar(ctx, service.EnviarRespuestaEntrada{ID: memoria.IDRespuestaEncuesta})
	assert.ErrorIs(t, err, service.ErrConflictoEstado)
}
