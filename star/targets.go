// SPDX-License-Identifier: GPL-3.0-only

package star

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target/local"
	"scampi.dev/scampi/target/rest"
	"scampi.dev/scampi/target/ssh"
)

// targetModule builds the `target` namespace (target.ssh, target.local).
func targetModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "target",
		Members: starlark.StringDict{
			"ssh":   starlark.NewBuiltin("target.ssh", builtinTargetSSH),
			"local": starlark.NewBuiltin("target.local", builtinTargetLocal),
			"rest":  starlark.NewBuiltin("target.rest", builtinTargetREST),
		},
	}
}

// target.ssh(name, host, user, ...)
func builtinTargetSSH(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var (
		name     string
		host     string
		user     string
		port     int
		key      string
		insecure bool
		timeout  string
	)
	if err := starlark.UnpackArgs("target.ssh", args, kwargs,
		"name", &name,
		"host", &host,
		"user", &user,
		"port?", &port,
		"key?", &key,
		"insecure?", &insecure,
		"timeout?", &timeout,
	); err != nil {
		return nil, err
	}

	span := callSpan(thread)
	pos := callerPosition(thread)
	call := findCallFromThread(thread, pos)

	if name == "" {
		s := span
		if call != nil {
			if vs, ok := kwargValueSpan(call, "name"); ok {
				s = vs
			}
		}
		return nil, &EmptyNameError{Func: "target.ssh", Source: s}
	}
	fields := kwargsFieldSpans(thread, "host", "user", "port", "key", "insecure", "timeout")

	inst := spec.TargetInstance{
		Type: ssh.SSH{},
		Config: &ssh.Config{
			Host: host, Port: port, User: user,
			Key: key, Insecure: insecure, Timeout: timeout,
		},
		Source: span,
		Fields: fields,
	}

	c := threadCollector(thread)
	if err := c.AddTarget(name, inst, span); err != nil {
		return nil, err
	}

	return poisonValue{funcName: "target.ssh"}, nil
}

// target.local(name)
func builtinTargetLocal(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs("target.local", args, kwargs,
		"name", &name,
	); err != nil {
		return nil, err
	}

	span := callSpan(thread)
	pos := callerPosition(thread)
	call := findCallFromThread(thread, pos)

	if name == "" {
		s := span
		if call != nil {
			if vs, ok := kwargValueSpan(call, "name"); ok {
				s = vs
			}
		}
		return nil, &EmptyNameError{Func: "target.local", Source: s}
	}
	inst := spec.TargetInstance{
		Type:   local.Local{},
		Config: &local.Config{},
		Source: span,
		Fields: make(map[string]spec.FieldSpan),
	}

	c := threadCollector(thread)
	if err := c.AddTarget(name, inst, span); err != nil {
		return nil, err
	}

	return poisonValue{funcName: "target.local"}, nil
}

// target.rest(name, base_url, auth?, tls?)
func builtinTargetREST(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var (
		name    string
		baseURL string
		authVal starlark.Value
		tlsVal  starlark.Value
	)
	if err := starlark.UnpackArgs("target.rest", args, kwargs,
		"name", &name,
		"base_url", &baseURL,
		"auth?", &authVal,
		"tls?", &tlsVal,
	); err != nil {
		return nil, err
	}

	span := callSpan(thread)
	pos := callerPosition(thread)
	call := findCallFromThread(thread, pos)

	if name == "" {
		s := span
		if call != nil {
			if vs, ok := kwargValueSpan(call, "name"); ok {
				s = vs
			}
		}
		return nil, &EmptyNameError{Func: "target.rest", Source: s}
	}

	cfg := &rest.Config{
		BaseURL: baseURL,
	}

	if authVal != nil && authVal != starlark.None {
		sa, ok := authVal.(starlarkAuth)
		if !ok {
			s := span
			if call != nil {
				if vs, ok := kwargValueSpan(call, "auth"); ok {
					s = vs
				}
			}
			return nil, &TypeError{
				Context:  "target.rest: auth",
				Expected: "rest.no_auth(), rest.basic(), rest.bearer(), or rest.header()",
				Got:      authVal.Type(),
				Source:   s,
			}
		}
		cfg.Auth = sa.config
	} else {
		cfg.Auth = rest.NoAuthConfig{}
	}

	if tlsVal != nil && tlsVal != starlark.None {
		st, ok := tlsVal.(starlarkTLS)
		if !ok {
			s := span
			if call != nil {
				if vs, ok := kwargValueSpan(call, "tls"); ok {
					s = vs
				}
			}
			return nil, &TypeError{
				Context:  "target.rest: tls",
				Expected: "rest.tls.secure(), rest.tls.insecure(), or rest.tls.ca_cert()",
				Got:      tlsVal.Type(),
				Source:   s,
			}
		}
		cfg.TLS = st.config
	} else {
		cfg.TLS = rest.SecureTLSConfig{}
	}

	fields := kwargsFieldSpans(thread, "base_url", "auth", "tls")
	inst := spec.TargetInstance{
		Type:   rest.REST{},
		Config: cfg,
		Source: span,
		Fields: fields,
	}

	c := threadCollector(thread)
	if err := c.AddTarget(name, inst, span); err != nil {
		return nil, err
	}

	return poisonValue{funcName: "target.rest"}, nil
}
