// SPDX-License-Identifier: GPL-3.0-only

package linker

import (
	"reflect"
	"strings"

	"scampi.dev/scampi/lang/eval"
)

// mapFields maps eval.Value fields onto a Go config struct pointer
// using reflection. Field names are matched by converting Go field
// names to snake_case.
func mapFields(fields map[string]eval.Value, cfg any) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()

	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := toSnake(f.Name)
		val, ok := fields[name]
		if !ok {
			continue
		}
		fv := v.Field(i)
		if err := setValue(fv, val); err != nil {
			return err
		}
	}
	return nil
}

// setValue assigns an eval.Value to a reflect.Value.
func setValue(dst reflect.Value, src eval.Value) error {
	if src == nil {
		return nil
	}
	switch sv := src.(type) {
	case *eval.StringVal:
		if dst.Kind() == reflect.String {
			dst.SetString(sv.V)
		}
	case *eval.IntVal:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dst.SetInt(sv.V)
		}
	case *eval.BoolVal:
		if dst.Kind() == reflect.Bool {
			dst.SetBool(sv.V)
		}
	case *eval.ListVal:
		if dst.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(dst.Type(), len(sv.Items), len(sv.Items))
			for i, item := range sv.Items {
				if err := setValue(slice.Index(i), item); err != nil {
					return err
				}
			}
			dst.Set(slice)
		}
	case *eval.MapVal:
		if dst.Kind() == reflect.Map {
			m := reflect.MakeMap(dst.Type())
			for i, k := range sv.Keys {
				kv := reflect.New(dst.Type().Key()).Elem()
				if err := setValue(kv, k); err != nil {
					return err
				}
				vv := reflect.New(dst.Type().Elem()).Elem()
				if err := setValue(vv, sv.Values[i]); err != nil {
					return err
				}
				m.SetMapIndex(kv, vv)
			}
			dst.Set(m)
		}
	case *eval.NoneVal:
		// Leave as zero value.
	case *eval.StructVal:
		// Nested struct value (e.g. Source, PkgSource composables).
		// These need special handling per target field type.
		// For now, store as-is if the destination is an interface.
		if dst.Kind() == reflect.Interface {
			dst.Set(reflect.ValueOf(sv))
		}
	}
	return nil
}

// toSnake converts GoFieldName to snake_case.
func toSnake(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
				if prev >= 'a' && prev <= 'z' {
					b.WriteByte('_')
				} else if prev >= 'A' && prev <= 'Z' && nextIsLower {
					b.WriteByte('_')
				}
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
