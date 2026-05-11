// Package schemagen generates type-safe schema definitions for GORM models.
//
// It parses GORM model fields via schema.Parse and generates a corresponding
// typed schema dictionary covering both columns and associations. These
// identifiers can be used to build query conditions, preloads, and joins with
// a type-safe SQL-building experience.
//
// It supports a custom naming strategy and allows configuring the output package
// name. By default, it generates one file per model in the output directory.
package schemagen

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"sync"

	"github.com/im-wmkong/gorm-query/internal/fsx"
	"github.com/im-wmkong/gorm-query/internal/reflectx"
	"github.com/im-wmkong/gorm-query/internal/slicex"
	"gorm.io/gorm/schema"
)

type Generator struct {
	cfg config
}

// New creates a new Generator.
//
// Example:
//
//	g := schemagen.New(
//	    schemagen.WithOutputDir("schema"),
//	)
//	_ = g
func New(opts ...Option) *Generator {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Generator{cfg: cfg}
}

// Generate parses the given GORM models and generates schema files into outputDir.
//
// Example:
//
//	// Generate schema into the default ./schema directory.
//	err := schemagen.New().Generate(&model.User{})
//	_ = err
func (g *Generator) Generate(models ...any) error {
	models, nilCount := g.filterNilModels(models)
	if nilCount > 0 {
		g.cfg.logger.Debug("filtered %d nil model(s), %d remaining", nilCount, len(models))
	}

	if len(models) == 0 {
		return fmt.Errorf("no models provided")
	}

	if err := g.checkSamePackage(models); err != nil {
		return err
	}

	pkgName, err := g.packageName()
	if err != nil {
		return err
	}

	if err := g.checkOutputDir(pkgName); err != nil {
		return err
	}

	g.cfg.logger.Info("generating schema for %d model(s)", len(models))

	schemas, err := g.schemas(models)
	if err != nil {
		return err
	}

	for _, sch := range schemas {
		content, err := newRenderer(g.cfg, sch, pkgName).Render()
		if err != nil {
			return err
		}

		filename := filepath.Join(g.cfg.outputDir, g.outputFile(sch))
		if err := g.writeFile(filename, content); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) filterNilModels(models []any) ([]any, int) {
	filtered := slicex.Filter(models, func(model any) bool {
		return model != nil
	})
	return filtered, len(models) - len(filtered)
}

func (g *Generator) checkSamePackage(models []any) error {
	if len(models) <= 1 {
		return nil
	}
	pkgName, ok := reflectx.PackageName(models[0])
	if !ok {
		return fmt.Errorf("model %T has no package path (anonymous or built-in types are not supported)", models[0])
	}
	for _, model := range models[1:] {
		currentPkg, ok := reflectx.PackageName(model)
		if !ok {
			return fmt.Errorf("model %T has no package path (anonymous or built-in types are not supported)", model)
		}
		if currentPkg != pkgName {
			return fmt.Errorf("all models must be in the same package, but found %q and %q", pkgName, currentPkg)
		}
	}
	return nil
}

func (g *Generator) packageName() (string, error) {
	if g.cfg.packageName != "" {
		g.cfg.logger.Debug("using explicit package name: %s", g.cfg.packageName)
		return g.cfg.packageName, nil
	}

	if outputPkg := fsx.InferPackageNameFromPath(g.cfg.outputDir); outputPkg != "" {
		g.cfg.logger.Debug("inferred package name from output dir: %s", outputPkg)
		return outputPkg, nil
	}

	return "", fmt.Errorf(
		"unable to resolve package name: output dir %q is not a valid Go identifier, please set WithPackageName explicitly",
		g.cfg.outputDir,
	)
}

func (g *Generator) checkOutputDir(pkgName string) error {
	if err := os.MkdirAll(g.cfg.outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir failed: %w", err)
	}

	fsPkg, err := fsx.ReadPackageNameFromDir(g.cfg.outputDir)
	if err != nil {
		return fmt.Errorf("detect package failed: %w", err)
	}
	if fsPkg != "" && fsPkg != pkgName {
		return fmt.Errorf("package mismatch: outputDir has package %q, but generator uses %q", fsPkg, pkgName)
	}

	return nil
}

func (g *Generator) schemas(models []any) ([]*schema.Schema, error) {
	seen := make(map[string]struct{})
	schemaCache := &sync.Map{}
	schemas := make([]*schema.Schema, 0, len(models))

	for _, model := range models {
		sch, err := schema.Parse(model, schemaCache, g.cfg.namingStrategy)
		if err != nil {
			return nil, fmt.Errorf("gorm schema parse failed: %w", err)
		}

		if sch.ModelType == nil || sch.ModelType.PkgPath() == "" || !token.IsExported(sch.ModelType.Name()) {
			g.cfg.logger.Warn("skipping unexported model: %s", sch.Name)
			continue
		}

		if len(sch.Fields) == 0 {
			g.cfg.logger.Warn("skipping model %s: no schema fields", sch.Name)
			continue
		}

		// Deduplicate by model name.
		if _, ok := seen[sch.Name]; ok {
			g.cfg.logger.Warn("skipping duplicate model: %s", sch.Name)
			continue
		}
		seen[sch.Name] = struct{}{}

		g.cfg.logger.Debug("parsed model %s: %d field(s)", sch.Name, len(sch.Fields))
		schemas = append(schemas, sch)
	}

	return schemas, nil
}

func (g *Generator) outputFile(sch *schema.Schema) string {
	return g.cfg.namingStrategy.ColumnName("", sch.Name) + "_gen.go"
}

func (g *Generator) writeFile(filename string, content []byte) error {
	if g.cfg.dryRun {
		return g.verifyDryRun(filename, content)
	}
	return g.writeIfChanged(filename, content)
}

// verifyDryRun checks that filename already contains the expected content.
func (g *Generator) verifyDryRun(filename string, content []byte) error {
	old, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("generated file missing: %s", filename)
		}
		return fmt.Errorf("read existing file %s failed: %w", filename, err)
	}
	if !bytes.Equal(old, content) {
		return fmt.Errorf("generated file outdated: %s", filename)
	}
	g.cfg.logger.Info("dry-run passed: %s is up to date", filename)
	return nil
}

// writeIfChanged writes content to filename, skipping the write when the
// existing file already matches.
func (g *Generator) writeIfChanged(filename string, content []byte) error {
	old, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing file %s failed: %w", filename, err)
	}
	if err == nil && bytes.Equal(old, content) {
		g.cfg.logger.Debug("skipped writing %s: content unchanged", filename)
		return nil
	}

	if err := os.WriteFile(filename, content, 0644); err != nil {
		return fmt.Errorf("write file %s failed: %w", filename, err)
	}

	g.cfg.logger.Info("generated %s", filename)
	return nil
}
