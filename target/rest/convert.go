// SPDX-License-Identifier: GPL-3.0-only

package rest

import (
	"reflect"

	"scampi.dev/scampi/lang/eval"
	"scampi.dev/scampi/spec"
)

// Converters returns the type converters owned by the REST target.
func Converters() spec.ConverterMap {
	return spec.ConverterMap{
		reflect.TypeFor[AuthConfig](): ConvertAuth,
		reflect.TypeFor[TLSConfig]():  ConvertTLS,
	}
}

// ConvertAuth converts a StructVal produced by rest.no_auth, rest.basic,
// rest.bearer, or rest.header into a rest.AuthConfig.
func ConvertAuth(typeName string, fields map[string]eval.Value, _ spec.ConvertContext) (any, error) {
	switch typeName {
	case "no_auth":
		return NoAuthConfig{}, nil
	case "basic":
		cfg := BasicAuthConfig{}
		if u, ok := fields["user"].(*eval.StringVal); ok {
			cfg.User = u.V
		}
		if p, ok := fields["password"].(*eval.StringVal); ok {
			cfg.Password = p.V
		}
		return cfg, nil
	case "bearer":
		cfg := BearerAuthConfig{}
		if t, ok := fields["token_endpoint"].(*eval.StringVal); ok {
			cfg.TokenEndpoint = t.V
		}
		if i, ok := fields["identity"].(*eval.StringVal); ok {
			cfg.Identity = i.V
		}
		if s, ok := fields["secret"].(*eval.StringVal); ok {
			cfg.Secret = s.V
		}
		return cfg, nil
	case "header":
		cfg := HeaderAuthConfig{}
		if n, ok := fields["name"].(*eval.StringVal); ok {
			cfg.Name = n.V
		}
		if v, ok := fields["value"].(*eval.StringVal); ok {
			cfg.Value = v.V
		}
		return cfg, nil
	}
	return NoAuthConfig{}, nil
}

// ConvertTLS converts a StructVal produced by rest.tls.secure or
// rest.tls.insecure into a rest.TLSConfig.
func ConvertTLS(typeName string, _ map[string]eval.Value, _ spec.ConvertContext) (any, error) {
	switch typeName {
	case "tls_insecure":
		return InsecureTLSConfig{}, nil
	}
	return SecureTLSConfig{}, nil
}
