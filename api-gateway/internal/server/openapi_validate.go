package server

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

func ValidateOpenAPI(path string) error {
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("load openapi: %w", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		return fmt.Errorf("validate openapi: %w", err)
	}
	return nil
}

func MustValidateOpenAPI(path string) {
	if err := ValidateOpenAPI(path); err != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI validation failed: %v\n", err)
		os.Exit(1)
	}
}
