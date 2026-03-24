// SPDX-License-Identifier: GPL-3.0-only

package star

import (
	"crypto/x509"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"scampi.dev/scampi/target/rest"
)

// restModule builds the `rest` namespace (rest.basic, rest.bearer, rest.header).
func restModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "rest",
		Members: starlark.StringDict{
			"no_auth": starlark.NewBuiltin("rest.no_auth", builtinRestNoAuth),
			"basic":   starlark.NewBuiltin("rest.basic", builtinRestBasic),
			"bearer":  starlark.NewBuiltin("rest.bearer", builtinRestBearer),
			"header":  starlark.NewBuiltin("rest.header", builtinRestHeader),
			"tls":     restTLSModule(),
		},
	}
}

// starlarkAuth wraps an AuthConfig so it can be passed through Starlark as a value.
type starlarkAuth struct {
	config rest.AuthConfig
}

func (a starlarkAuth) String() string        { return "<rest.auth:" + a.config.Kind() + ">" }
func (a starlarkAuth) Type() string          { return "rest.auth" }
func (a starlarkAuth) Freeze()               {}
func (a starlarkAuth) Truth() starlark.Bool  { return starlark.True }
func (a starlarkAuth) Hash() (uint32, error) { return 0, nil }

// rest.no_auth()
func builtinRestNoAuth(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackArgs("rest.no_auth", args, kwargs); err != nil {
		return nil, err
	}
	return starlarkAuth{config: rest.NoAuthConfig{}}, nil
}

// rest.basic(user, password)
func builtinRestBasic(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var user, password string
	if err := starlark.UnpackArgs("rest.basic", args, kwargs,
		"user", &user,
		"password", &password,
	); err != nil {
		return nil, err
	}
	return starlarkAuth{config: rest.BasicAuthConfig{User: user, Password: password}}, nil
}

// rest.bearer(token_endpoint, identity, secret)
func builtinRestBearer(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var tokenEndpoint, identity, secret string
	if err := starlark.UnpackArgs("rest.bearer", args, kwargs,
		"token_endpoint", &tokenEndpoint,
		"identity", &identity,
		"secret", &secret,
	); err != nil {
		return nil, err
	}
	return starlarkAuth{config: rest.BearerAuthConfig{
		TokenEndpoint: tokenEndpoint,
		Identity:      identity,
		Secret:        secret,
	}}, nil
}

// TLS
// -----------------------------------------------------------------------------

func restTLSModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "rest.tls",
		Members: starlark.StringDict{
			"secure":   starlark.NewBuiltin("rest.tls.secure", builtinRestTLSSecure),
			"insecure": starlark.NewBuiltin("rest.tls.insecure", builtinRestTLSInsecure),
			"ca_cert":  starlark.NewBuiltin("rest.tls.ca_cert", builtinRestTLSCACert),
		},
	}
}

type starlarkTLS struct {
	config rest.TLSConfig
}

func (s starlarkTLS) String() string        { return "<rest.tls:" + s.config.Kind() + ">" }
func (s starlarkTLS) Type() string          { return "rest.tls" }
func (s starlarkTLS) Freeze()               {}
func (s starlarkTLS) Truth() starlark.Bool  { return starlark.True }
func (s starlarkTLS) Hash() (uint32, error) { return 0, nil }

func builtinRestTLSSecure(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackArgs("rest.tls.secure", args, kwargs); err != nil {
		return nil, err
	}
	return starlarkTLS{config: rest.SecureTLSConfig{}}, nil
}

func builtinRestTLSInsecure(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackArgs("rest.tls.insecure", args, kwargs); err != nil {
		return nil, err
	}
	return starlarkTLS{config: rest.InsecureTLSConfig{}}, nil
}

func builtinRestTLSCACert(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs("rest.tls.ca_cert", args, kwargs,
		"path", &path,
	); err != nil {
		return nil, err
	}

	span := callSpan(thread)
	c := threadCollector(thread)

	pem, err := c.src.ReadFile(c.ctx, path)
	if err != nil {
		return nil, &CACertReadError{Path: path, Source: span, Err: err}
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, &CACertParseError{Path: path, Source: span}
	}

	return starlarkTLS{config: rest.CACertTLSConfig{Pool: pool}}, nil
}

// rest.header(name, value)
func builtinRestHeader(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var name, value string
	if err := starlark.UnpackArgs("rest.header", args, kwargs,
		"name", &name,
		"value", &value,
	); err != nil {
		return nil, err
	}
	return starlarkAuth{config: rest.HeaderAuthConfig{Name: name, Value: value}}, nil
}
