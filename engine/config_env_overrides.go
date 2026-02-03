package engine

import (
	"errors"
	"fmt"
	"strconv"

	"cuelang.org/go/cue"
	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/errs"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/source"
)

func extractEnvMap(v cue.Value) map[string]string {
	res := make(map[string]string)

	iter, _ := v.Fields(cue.Optional(true))
	for iter.Next() {
		field := iter.Selector().String()
		envVar := extractAttr(iter.Value(), cueAttrEnv)
		if envVar != "" {
			res[field] = envVar
		}
	}
	return res
}

func applyEnvOverrides(v cue.Value, src source.Source) (cue.Value, error) {
	envMap := extractEnvMap(v)
	var diags diagnostic.Diagnostics
	for field, envVar := range envMap {
		if envVal, ok := src.LookupEnv(envVar); ok {
			path := cue.ParsePath(field)
			kind := v.LookupPath(path).IncompleteKind()
			if val, diag := parseEnvVal(kind, envVar, envVal); diag != nil {
				diags = append(diags, diag)
			} else {
				v = v.FillPath(path, val)
			}
		}
	}

	if len(diags) == 0 {
		return v, nil
	}

	return v, diags
}

func parseEnvVal(kind cue.Kind, envVar, envVal string) (any, diagnostic.Diagnostic) {
	switch {
	case kind.IsAnyOf(cue.StringKind):
		return envVal, nil

	case kind.IsAnyOf(cue.IntKind):
		iVal, err := strconv.ParseUint(envVal, 0, 0)
		if err == nil {
			return iVal, nil
		}

		detectBase := func(s string) int {
			if len(s) < 2 {
				return 10
			}

			switch s[:2] {
			case "0x", "0X":
				return 16
			case "0b", "0B":
				return 2
			case "0o", "0O":
				return 8
			default:
				return 10
			}
		}

		kindStr := fmt.Sprintf("int (base %d)", detectBase(envVal))
		switch {
		case errors.Is(err, strconv.ErrSyntax):
			return nil, InvalidEnvVar{
				Key:   envVar,
				Value: envVal,
				Kind:  kindStr,
				Err:   errs.UnwrapAll(err),
			}
		case errors.Is(err, strconv.ErrRange):
			return nil, EnvVarOutOfRange{
				Key:   envVar,
				Value: envVal,
				Kind:  kindStr,
				Err:   errs.UnwrapAll(err),
			}
		}

	case kind.IsAnyOf(cue.FloatKind):
		fVal, err := strconv.ParseFloat(envVal, 64)
		if err == nil {
			return fVal, nil
		}

		return nil, InvalidEnvVar{
			Key:   envVar,
			Value: envVal,
			Kind:  "float",
			Err:   err,
		}

	case kind.IsAnyOf(cue.BoolKind):
		bVal, err := strconv.ParseBool(envVal)
		if err == nil {
			return bVal, nil
		}
		return nil, InvalidEnvVar{
			Key:   envVar,
			Value: envVal,
			Kind:  "bool",
			Err:   err,
		}
	}

	return nil, nil
}

type InvalidEnvVar struct {
	Key   string
	Value string
	Kind  string
	Err   error
}

func (e InvalidEnvVar) Error() string {
	return fmt.Sprintf("invalid environment variable %q (%q): %v", e.Key, e.Value, e.Err)
}

func (e InvalidEnvVar) EventTemplate() event.Template {
	return event.Template{
		ID:   "env.InvalidEnvVar",
		Text: `failed to parse ENV "{{.Key}}"`,
		Hint: `"{{.Value}}" could not be parsed to {{.Kind}}`,
		Help: `{{- if .Err}}underlying error was: {{.Err}}{{end}}`,
		Data: e,
	}
}

func (InvalidEnvVar) Severity() signal.Severity { return signal.Warning }
func (InvalidEnvVar) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type EnvVarOutOfRange struct {
	Key   string
	Value string
	Kind  string
	Err   error
}

func (e EnvVarOutOfRange) Error() string {
	return fmt.Sprintf("environment variable out of range %q (%q): %v", e.Key, e.Value, e.Err)
}

func (e EnvVarOutOfRange) EventTemplate() event.Template {
	return event.Template{
		ID:   "env.EnvVarOutOfRange",
		Text: `failed to parse ENV "{{.Key}}"`,
		Hint: `"{{.Value}}" is out of range for {{.Kind}}`,
		Help: `{{- if .Err}}underlying error was: {{.Err}}{{end}}`,
		Data: e,
	}
}

func (EnvVarOutOfRange) Severity() signal.Severity { return signal.Warning }
func (EnvVarOutOfRange) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
