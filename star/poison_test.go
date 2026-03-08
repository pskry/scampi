// SPDX-License-Identifier: GPL-3.0-only

package star

import (
	"testing"

	"go.starlark.net/starlark"
)

func TestCheckPoison(t *testing.T) {
	poison := poisonValue{funcName: "deploy"}

	tests := []struct {
		name    string
		val     starlark.Value
		wantErr bool
	}{
		{
			name:    "direct poison",
			val:     poison,
			wantErr: true,
		},
		{
			name: "poison in list",
			val: func() *starlark.List {
				l := starlark.NewList([]starlark.Value{starlark.String("ok"), poison})
				return l
			}(),
			wantErr: true,
		},
		{
			name: "poison in dict value",
			val: func() *starlark.Dict {
				d := new(starlark.Dict)
				_ = d.SetKey(starlark.String("key"), poison)
				return d
			}(),
			wantErr: true,
		},
		{
			name:    "poison in tuple",
			val:     starlark.Tuple{starlark.String("ok"), poison},
			wantErr: true,
		},
		{
			name:    "clean string",
			val:     starlark.String("hello"),
			wantErr: false,
		},
		{
			name:    "clean int",
			val:     starlark.MakeInt(42),
			wantErr: false,
		},
		{
			name: "clean list of strings",
			val: func() *starlark.List {
				return starlark.NewList([]starlark.Value{starlark.String("a"), starlark.String("b")})
			}(),
			wantErr: false,
		},
		{
			name: "clean dict",
			val: func() *starlark.Dict {
				d := new(starlark.Dict)
				_ = d.SetKey(starlark.String("k"), starlark.String("v"))
				return d
			}(),
			wantErr: false,
		},
		{
			name:    "empty list",
			val:     starlark.NewList(nil),
			wantErr: false,
		},
		{
			name:    "empty dict",
			val:     new(starlark.Dict),
			wantErr: false,
		},
		{
			name:    "empty tuple",
			val:     starlark.Tuple{},
			wantErr: false,
		},
		{
			name: "deeply nested poison",
			val: func() *starlark.List {
				inner := starlark.NewList([]starlark.Value{poison})
				outer := starlark.NewList([]starlark.Value{inner})
				return outer
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPoison(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkPoison() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				pe, ok := err.(*PoisonValueError)
				if !ok {
					t.Errorf("expected *PoisonValueError, got %T", err)
				} else if pe.FuncName != "deploy" {
					t.Errorf("FuncName = %q, want %q", pe.FuncName, "deploy")
				}
			}
		})
	}
}
