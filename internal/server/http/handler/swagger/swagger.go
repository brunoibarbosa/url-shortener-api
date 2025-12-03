package http_handler

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type SwaggerHandler struct {
	specPath string
	baseDir  string
}

func NewSwaggerHandler(specPath string) *SwaggerHandler {
	absSpecPath, err := filepath.Abs(specPath)
	if err != nil {
		log.Printf("Warning: Could not get absolute path for %s: %v", specPath, err)
		absSpecPath = specPath
	}

	baseDir := filepath.Dir(absSpecPath)

	log.Printf("Swagger: specPath=%s, absPath=%s, baseDir=%s", specPath, absSpecPath, baseDir)

	if _, err := os.Stat(absSpecPath); os.IsNotExist(err) {
		log.Printf("Warning: OpenAPI spec file not found at %s", absSpecPath)
	} else {
		log.Printf("OpenAPI spec file found at %s", absSpecPath)
	}

	return &SwaggerHandler{
		specPath: absSpecPath,
		baseDir:  baseDir,
	}
}

func (h *SwaggerHandler) Handler() http.Handler {
	if _, err := os.Stat(h.specPath); os.IsNotExist(err) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "OpenAPI specification file not found: "+h.specPath, http.StatusNotFound)
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
	h.serveYAMLFile(w, r, h.specPath)
	return nil
}

func (h *SwaggerHandler) ServeSpecFiles(w http.ResponseWriter, r *http.Request) {
	requestedPath := strings.TrimPrefix(r.URL.Path, "/docs/")

	if requestedPath == "" || requestedPath == "/" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
	}

	requestedPath = filepath.FromSlash(requestedPath)
	filePath := filepath.Join(h.baseDir, requestedPath)

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	if !strings.HasPrefix(absFilePath, h.baseDir) {
		http.Error(w, "Invalid file path", http.StatusForbidden)
	}

	h.serveYAMLFile(w, r, absFilePath)
}

func (h *SwaggerHandler) serveYAMLFile(w http.ResponseWriter, _ *http.Request, filePath string) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "File not found: "+filepath.Base(filePath), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}
