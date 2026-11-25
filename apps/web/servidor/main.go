// Comando servidor: sirve la app web compilada (spa-server, DAS §4).
//
// Sirve los archivos de /srv, responde /healthz, agrega cabeceras de seguridad,
// cachea /assets/ como inmutable y reenvía /api al backend (BACKEND_URL), para que
// el navegador vea un solo origen y la cookie de sesión funcione sin CORS.
package main

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	backend := os.Getenv("BACKEND_URL")
	if backend == "" {
		slog.Error("falta la variable de entorno BACKEND_URL")
		os.Exit(1)
	}
	manejador, err := nuevoManejador("/srv", backend)
	if err != nil {
		slog.Error("configuración inválida", "error", err)
		os.Exit(1)
	}
	servidor := &http.Server{Addr: ":8080", Handler: manejador, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("servidor web escuchando", "puerto", 8080)
	if err := servidor.ListenAndServe(); err != nil {
		slog.Error("el servidor web terminó con error", "error", err)
		os.Exit(1)
	}
}

// nuevoManejador arma el router del servidor web sobre el directorio raiz.
func nuevoManejador(raiz, backendURL string) (http.Handler, error) {
	destino, err := url.Parse(backendURL)
	if err != nil || destino.Scheme == "" || destino.Host == "" {
		return nil, errors.New("BACKEND_URL debe ser una URL absoluta")
	}
	proxy := httputil.NewSingleHostReverseProxy(destino)
	archivos := http.FileServer(http.Dir(raiz))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"estado":"ok"}`))
	})
	mux.Handle("/api/", proxy)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		limpia := path.Clean("/" + r.URL.Path)
		if strings.HasPrefix(limpia, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			archivos.ServeHTTP(w, r)
			return
		}
		// Rutas de la SPA sin archivo propio: se entrega index.html.
		if info, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(limpia))); err != nil || info.IsDir() {
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, filepath.Join(raiz, "index.html"))
			return
		}
		archivos.ServeHTTP(w, r)
	})
	return conCabecerasSeguras(mux), nil
}

// conCabecerasSeguras agrega las cabeceras que pide el DAS a toda respuesta.
func conCabecerasSeguras(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		siguiente.ServeHTTP(w, r)
	})
}
