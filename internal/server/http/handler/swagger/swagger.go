package http_handler

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/brunoibarbosa/url-shortener/internal/openapi"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type SwaggerHandler struct {
	specPath    string
	baseDir     string
	bundledSpec []byte
	bundleError error
}

func NewSwaggerHandler(specPath string) *SwaggerHandler {
	absSpecPath, err := filepath.Abs(specPath)
	if err != nil {
		log.Printf("Warning: Could not get absolute path for %s: %v", specPath, err)
		absSpecPath = specPath
	}

	baseDir := filepath.Dir(absSpecPath)

	log.Printf("Swagger: specPath=%s, absPath=%s, baseDir=%s", specPath, absSpecPath, baseDir)

	handler := &SwaggerHandler{
		specPath: absSpecPath,
		baseDir:  baseDir,
	}

	if _, err := os.Stat(absSpecPath); os.IsNotExist(err) {
		log.Printf("Warning: OpenAPI spec file not found at %s", absSpecPath)
		handler.bundleError = err
		return handler
	}

	log.Printf("OpenAPI spec file found at %s", absSpecPath)

	log.Println("Bundling OpenAPI specification...")
	bundler := openapi.NewBundler(absSpecPath)
	bundled, err := bundler.Bundle(absSpecPath)
	if err != nil {
		log.Printf("Error bundling OpenAPI spec: %v", err)
		handler.bundleError = err
		return handler
	}

	handler.bundledSpec = bundled
	log.Printf("OpenAPI specification bundled successfully (%d bytes)", len(bundled))

	return handler
}

func (h *SwaggerHandler) Handler() http.Handler {
	if h.bundleError != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "OpenAPI specification error: "+h.bundleError.Error(), http.StatusInternalServerError)
		})
	}

	return httpSwagger.Handler(
		httpSwagger.URL("/docs/openapi.yaml"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)
}

func (h *SwaggerHandler) ServeSpec(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	if h.bundleError != nil {
		http.Error(w, "OpenAPI specification error: "+h.bundleError.Error(), http.StatusInternalServerError)
		return nil
	}

	w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write(h.bundledSpec)

	return nil
}
