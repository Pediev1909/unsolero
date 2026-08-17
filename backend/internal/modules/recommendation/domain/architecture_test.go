package domain

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRecommendationProductionCodeHasNoInfrastructureOrCommercialImports(t *testing.T) {
	patterns := []string{"*.go", "../application/*.go", "../ports/*.go"}
	files := make([]string, 0)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("find recommendation source files: %v", err)
		}
		files = append(files, matches...)
	}
	forbidden := []string{
		"/adapters/",
		"/transport/",
		"/modules/commerce",
		"/modules/analytics",
		"/modules/ai",
		"/integrations/ai",
	}
	for _, filename := range files {
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", filename, err)
		}
		for _, imported := range parsed.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatalf("read import in %s: %v", filename, err)
			}
			for _, blocked := range forbidden {
				if strings.Contains(path, blocked) {
					t.Fatalf("production recommendation file %s imports forbidden dependency %s", filename, path)
				}
			}
		}
	}
}
