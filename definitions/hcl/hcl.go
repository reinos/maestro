package hcl

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jexia/maestro/logger"
	"github.com/jexia/maestro/schema"
	"github.com/jexia/maestro/specs"
	"github.com/jexia/maestro/utils"
)

// SchemaResolver constructs a schema resolver for the given path.
// The HCL schema resolver relies on other schema registries.
// Those need to be resolved before the HCL schemas are resolved.
func SchemaResolver(path string) schema.Resolver {
	return func(ctx context.Context, schemas *schema.Store) error {
		files, err := utils.ResolvePath(path)
		if err != nil {
			return err
		}

		for _, file := range files {
			reader, err := os.Open(file.Path)
			if err != nil {
				return err
			}

			definition, err := UnmarshalHCL(ctx, file.Name(), reader)
			if err != nil {
				return err
			}

			collection, err := ParseSchema(ctx, definition, schemas)
			if err != nil {
				return err
			}

			schemas.Add(collection)
		}

		return nil
	}
}

// DefinitionResolver constructs a definition resolver for the given path
func DefinitionResolver(path string) specs.Resolver {
	return func(ctx context.Context, functions specs.CustomDefinedFunctions) (*specs.Manifest, error) {
		files, err := utils.ResolvePath(path)
		if err != nil {
			return nil, err
		}

		result := &specs.Manifest{}

		for _, file := range files {
			reader, err := os.Open(file.Path)
			if err != nil {
				return nil, err
			}

			definition, err := UnmarshalHCL(ctx, file.Name(), reader)
			if err != nil {
				return nil, err
			}

			manifest, err := ParseSpecs(ctx, definition, functions)
			if err != nil {
				return nil, err
			}

			result.Merge(manifest)
		}

		return result, nil
	}
}

// UnmarshalHCL unmarshals the given HCL stream into a intermediate resource.
func UnmarshalHCL(ctx context.Context, filename string, reader io.Reader) (manifest Manifest, _ error) {
	logger.FromCtx(ctx, logger.Core).WithField("file", filename).Info("Reading HCL files")

	bb, err := ioutil.ReadAll(reader)
	if err != nil {
		return manifest, err
	}

	logger.FromCtx(ctx, logger.Core).WithField("file", filename).Debug("Parsing HCL syntax")

	file, diags := hclsyntax.ParseConfig(bb, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return manifest, errors.New(diags.Error())
	}

	logger.FromCtx(ctx, logger.Core).WithField("file", filename).Debug("Decoding HCL syntax")

	diags = gohcl.DecodeBody(file.Body, nil, &manifest)
	if diags.HasErrors() {
		return manifest, errors.New(diags.Error())
	}

	return manifest, nil
}
