package domain

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAIModuleHasNoInfrastructureDependencies(t *testing.T) {
	patterns := []string{"*.go", "../application/*.go", "../ports/*.go", "../../../adapters/ai/*.go"}
	forbidden := []string{"/adapters/", "/platform/database", "database/sql", "github.com/jackc/pgx"}
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("find AI source files: %v", err)
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
						t.Fatalf("AI production file %s imports forbidden dependency %s", filename, path)
					}
				}
			}
		}
	}
}
