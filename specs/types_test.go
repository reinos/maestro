package specs

import (
	"context"
	"math/big"
	"testing"

	"github.com/jexia/maestro/logger"
	"github.com/jexia/maestro/specs/types"
	"github.com/zclconf/go-cty/cty"
)

func TestSetDefaultValue(t *testing.T) {
	type expected struct {
		Default interface{}
		Type    types.Type
	}

	tests := map[cty.Value]expected{
		cty.StringVal("default"): {
			Default: "default",
			Type:    types.TypeString,
		},
		cty.NumberVal(big.NewFloat(10)): {
			Default: int64(10),
			Type:    types.TypeInt64,
		},
		cty.BoolVal(true): {
			Default: true,
			Type:    types.TypeBool,
		},
	}

	for input, expected := range tests {
		ctx := context.Background()
		ctx = logger.WithValue(ctx)

		property := Property{}
		SetDefaultValue(ctx, &property, input)

		if expected.Default != property.Default {
			t.Errorf("unexpected result %+v, expected %+v", property.Default, expected.Default)
		}

		if expected.Type != property.Type {
			t.Errorf("unexpected type %s, expected %s", property.Type, expected.Type)
		}
	}
}
