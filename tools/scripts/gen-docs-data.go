package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FieldInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Tag     string `json:"tag"`
	Comment string `json:"comment"`
}

type StructInfo struct {
	Name    string      `json:"name"`
	Doc     string      `json:"doc"`
	Fields  []FieldInfo `json:"fields"`
	Package string      `json:"package"`
}

type DocManifest struct {
	GeneratedAt string       `json:"generated_at"`
	Structs     []StructInfo `json:"structs"`
}

func main() {
	paths := []string{
		"services/refinery/pkg/config",
		"services/refinery/pkg/audit",
		"services/refinery/pkg/license",
	}

	manifest := DocManifest{
		GeneratedAt: fmt.Sprintf("%v", os.Getenv("TIME")), // Placeholder or use real time
	}

	for _, p := range paths {
		structs, err := parsePackage(p)
		if err != nil {
			log.Printf("[WARN] Failed to parse %s: %v", p, err)
			continue
		}
		manifest.Structs = append(manifest.Structs, structs...)
	}

	// Ensure output directory exists
	outDir := "apps/web/src/docs"
	os.MkdirAll(outDir, 0755)

	outFile := filepath.Join(outDir, "docs-manifest.json")
	data, _ := json.MarshalIndent(manifest, "", "  ")
	err := os.WriteFile(outFile, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write manifest: %v", err)
	}

	fmt.Printf("Successfully generated doc manifest at %s (%d structs)\n", outFile, len(manifest.Structs))
}

func parsePackage(path string) ([]StructInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var results []StructInfo
	for pkgName, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return true
				}

				info := StructInfo{
					Name:    ts.Name.Name,
					Package: pkgName,
					Doc:     strings.TrimSpace(ts.Doc.Text()),
				}

				for _, field := range st.Fields.List {
					var fieldNames []string
					for _, name := range field.Names {
						fieldNames = append(fieldNames, name.Name)
					}

					typeName := fmt.Sprintf("%v", field.Type) // Simplified
					tag := ""
					if field.Tag != nil {
						tag = strings.Trim(field.Tag.Value, "`")
					}

					comment := strings.TrimSpace(field.Doc.Text())
					if comment == "" {
						comment = strings.TrimSpace(field.Comment.Text())
					}

					info.Fields = append(info.Fields, FieldInfo{
						Name:    strings.Join(fieldNames, ", "),
						Type:    typeName,
						Tag:     tag,
						Comment: comment,
					})
				}

				results = append(results, info)
				return true
			})
		}
	}

	return results, nil
}
