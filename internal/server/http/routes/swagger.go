package http_routes

import (
	"os"
	"path/filepath"

	"github.com/brunoibarbosa/url-shortener/internal/server/http"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler/swagger"
)

type SwaggerRoutesConfig struct {
	SpecPath string
}

func NewSwaggerRoutes(r *http.AppRouter, config SwaggerRoutesConfig) {
	specPath := config.SpecPath
	if !filepath.IsAbs(specPath) {
		executable, err := os.Executable()
		var projectRoot string
		if err == nil {
			projectRoot = filepath.Join(filepath.Dir(executable), "..", "..")
		} else {
			wd, _ := os.Getwd()
			if filepath.Base(filepath.Dir(wd)) == "cmd" {
				projectRoot = filepath.Join(wd, "..", "..")
			} else {
				projectRoot = wd
			}
		}
		specPath = filepath.Join(projectRoot, config.SpecPath)
	}

	handler := http_handler.NewSwaggerHandler(specPath)

	r.Get("/docs/openapi.yaml", handler.ServeSpec)

	r.HandleFunc("/docs/*", handler.ServeSpecFiles)

	r.HandleFunc("/swagger/*", handler.Handler().ServeHTTP)
	r.HandleFunc("/swagger", handler.Handler().ServeHTTP)
}
