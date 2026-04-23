// SPDX-License-Identifier: GPL-3.0-only

package rest

import (
	"reflect"

	"github.com/itchyny/gojq"

	"scampi.dev/scampi/lang/eval"
	"scampi.dev/scampi/spec"
)

// Converters returns the type converters owned by the REST step.
func Converters() spec.ConverterMap {
	return spec.ConverterMap{
		reflect.TypeFor[BodyConfig]():  ConvertBody,
		reflect.TypeFor[CheckConfig](): ConvertCheck,
		reflect.TypeFor[*JQBinding]():  ConvertBinding,
	}
}

// ConvertBody converts a StructVal produced by rest.body.json or
// rest.body.string into a BodyConfig.
func ConvertBody(typeName string, fields map[string]eval.Value, _ spec.ConvertContext) (any, error) {
	switch typeName {
	case "body_json":
		if d, ok := fields["data"]; ok {
			return JSONBody{Data: toGo(d)}, nil
		}
		return JSONBody{}, nil
	case "body_string":
		if c, ok := fields["content"].(*eval.StringVal); ok {
			return StringBody{Content: c.V}, nil
		}
		return StringBody{}, nil
	}
	return nil, nil
}

// ConvertBinding converts a StructVal into a *JQBinding.
func ConvertBinding(_ string, fields map[string]eval.Value, _ spec.ConvertContext) (any, error) {
	b := &JQBinding{}
	if e, ok := fields["expr"].(*eval.StringVal); ok {
		b.Expr = e.V
		q, err := gojq.Parse(e.V)
		if err == nil {
			compiled, err := gojq.Compile(q)
			if err == nil {
				b.Compiled = compiled
			}
		}
	}
	return b, nil
}

// ConvertCheck converts a StructVal produced by rest.status or rest.jq
// into a CheckConfig.
func ConvertCheck(typeName string, fields map[string]eval.Value, _ spec.ConvertContext) (any, error) {
	switch typeName {
	case "status":
		c := StatusCheck{}
		if code, ok := fields["code"].(*eval.IntVal); ok {
			c.Status = int(code.V)
		}
		return c, nil
	case "jq":
		c := &JQCheck{}
		if e, ok := fields["expr"].(*eval.StringVal); ok {
			c.Expr = e.V
			q, err := gojq.Parse(e.V)
			if err == nil {
				compiled, err := gojq.Compile(q)
				if err == nil {
					c.Compiled = compiled
				}
			}
		}
		return c, nil
	}
	return nil, nil
}

// toGo converts an eval.Value to a Go native type.
func toGo(v eval.Value) any {
	switch sv := v.(type) {
	case *eval.StringVal:
		return sv.V
	case *eval.IntVal:
		return sv.V
	case *eval.BoolVal:
		return sv.V
	case *eval.NoneVal:
		return nil
	case *eval.ListVal:
		r := make([]any, len(sv.Items))
		for i, item := range sv.Items {
			r[i] = toGo(item)
		}
		return r
	case *eval.MapVal:
		m := make(map[string]any, len(sv.Keys))
		for i, k := range sv.Keys {
			if sk, ok := k.(*eval.StringVal); ok {
				m[sk.V] = toGo(sv.Values[i])
			}
		}
		return m
	}
	return nil
}
