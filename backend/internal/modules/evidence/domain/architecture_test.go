package domain

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestEvidenceProductionCodeHasNoAIOrCommercialImports(t *testing.T) {
	patterns := []string{"*.go", "../application/*.go", "../ports/*.go"}
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("find evidence source files: %v", err)
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
				for _, forbidden := range []string{"/modules/ai", "/modules/commerce", "/modules/analytics"} {
					if strings.Contains(path, forbidden) {
						t.Fatalf("production evidence file %s imports forbidden dependency %s", filename, path)
					}
				}
			}
		}
	}
}
