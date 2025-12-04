package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/brunoibarbosa/url-shortener/internal/openapi"
)

func main() {
	inputPath := flag.String("input", "docs/openapi/openapi.yaml", "Caminho para o arquivo OpenAPI de entrada")
	outputPath := flag.String("output", "docs/openapi.bundled.yaml", "Caminho para o arquivo OpenAPI consolidado de saída")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	if filepath.Base(wd) == "bundle-openapi" && filepath.Base(filepath.Dir(wd)) == "cmd" {
		wd = filepath.Join(wd, "..", "..")
	}

	absInputPath := filepath.Join(wd, *inputPath)
	absOutputPath := filepath.Join(wd, *outputPath)

	log.Printf("Bundling OpenAPI specification...")
	log.Printf("Input:  %s", absInputPath)
	log.Printf("Output: %s", absOutputPath)

	bundler := openapi.NewBundler(absInputPath)

	if err := bundler.BundleToFile(absInputPath, absOutputPath); err != nil {
		log.Fatalf("Failed to bundle OpenAPI spec: %v", err)
	}

	info, err := os.Stat(absOutputPath)
	if err != nil {
		log.Fatalf("Failed to stat output file: %v", err)
	}

	log.Printf("✅ OpenAPI specification bundled successfully!")
	log.Printf("📄 Output file: %s (%d bytes)", absOutputPath, info.Size())
}
