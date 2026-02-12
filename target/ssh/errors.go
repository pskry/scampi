// SPDX-License-Identifier: GPL-3.0-only

package ssh

import (
	"encoding/base64"
	"fmt"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
	"golang.org/x/crypto/ssh/knownhosts"
)

type NoKnownHostsError struct {
	Path string
	Err  error
}

func (e NoKnownHostsError) Error() string {
	return fmt.Sprintf("known_hosts file not found: %s", e.Path)
}

func (e NoKnownHostsError) Unwrap() error { return e.Err }

func (e NoKnownHostsError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.NoKnownHosts",
		Text: `known_hosts file "{{.Path}}" not found`,
		Hint: "create the file or use insecure: true to skip host key verification",
		Help: "without a known_hosts file, host key verification cannot proceed",
		Data: e,
	}
}

func (NoKnownHostsError) Severity() signal.Severity { return signal.Error }
func (NoKnownHostsError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type NoSuchHostError struct {
	Host   string
	Source spec.SourceSpan
}

func (e NoSuchHostError) Error() string {
	return fmt.Sprintf("no such host %s", e.Host)
}

func (e NoSuchHostError) EventTemplate() event.Template {
	return event.Template{
		ID:     "ssh.NoSuchHost",
		Text:   "no such host {{.Host}}",
		Hint:   "make sure the host is reachable",
		Source: &e.Source,
		Data:   e,
	}
}

func (NoSuchHostError) Severity() signal.Severity { return signal.Error }
func (NoSuchHostError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type ConnectionError struct {
	Host string
	Port int
	Err  error
}

func (e ConnectionError) Error() string {
	return fmt.Sprintf("failed to connect to %s:%d: %v", e.Host, e.Port, e.Err)
}

func (e ConnectionError) Unwrap() error { return e.Err }

func (e ConnectionError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.Connection",
		Text: "failed to connect to {{.Host}}:{{.Port}}",
		Hint: "make sure the host is reachable and SSH is running on the given port",
		Help: "underlying error was: {{.Err}}",
		Data: e,
	}
}

func (ConnectionError) Severity() signal.Severity { return signal.Error }
func (ConnectionError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type UnknownKeyError struct {
	Err error
}

func (e UnknownKeyError) Error() string {
	return fmt.Sprintf("unknown host SSH-key: %v", e.Err)
}

func (e UnknownKeyError) Unwrap() error { return e.Err }

func (e UnknownKeyError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.UnknownKey",
		Text: "unknown host SSH-key",
		Hint: "make sure the host SSH-key is known",
	}
}

func (UnknownKeyError) Severity() signal.Severity { return signal.Error }
func (UnknownKeyError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

func toKnownKeys(keys []knownhosts.KnownKey) []KnownKey {
	l := len(keys)
	res := make([]KnownKey, l)
	for i := range l {
		res[i] = toKnownKey(keys[i])
	}
	return res
}

func toKnownKey(k knownhosts.KnownKey) KnownKey {
	fingerprint := func(s string) string {
		l := len(s)
		if l <= 7*2 {
			return s
		}

		return s[:7] + "..." + s[l-7:]
	}
	key := base64.StdEncoding.EncodeToString(k.Key.Marshal())

	return KnownKey{
		Type:        k.Key.Type(),
		Key:         key,
		Fingerprint: fingerprint(key),
		Filename:    k.Filename,
		Line:        k.Line,
	}
}

type KnownKey struct {
	Type        string
	Key         string
	Fingerprint string
	Filename    string
	Line        int
}

type KeyMismatchError struct {
	Known []KnownKey
	Err   error
}

func (e KeyMismatchError) Error() string {
	return fmt.Sprintf("host SSH-key mismatch: %v", e.Err)
}

func (e KeyMismatchError) Unwrap() error { return e.Err }

func (e KeyMismatchError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.KeyMismatch",
		Text: "host SSH-key mismatch",
		Hint: "make sure the host SSH-key is correct",
		Help: `known host keys:
{{- range .Known}}
  - {{.Filename}}:{{.Line}}: {{.Type}} {{.Fingerprint}}
{{end}}`,
		Data: e,
	}
}

func (KeyMismatchError) Severity() signal.Severity { return signal.Error }
func (KeyMismatchError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type KeyRevokedError struct {
	Revoked KnownKey
	Err     error
}

func (e KeyRevokedError) Error() string {
	return fmt.Sprintf("host SSH-key revoked: %v", e.Err)
}

func (e KeyRevokedError) Unwrap() error { return e.Err }

func (e KeyRevokedError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.KeyRevoked",
		Text: "host SSH-key revoked",
		Hint: "make sure the host SSH-key is up-to-date",
		Help: `revoked host key:
  {{.Revoked.Filename}}:{{.Revoked.Line}}: {{.Revoked.Type}} {{.Revoked.Fingerprint}}`,
		Data: e,
	}
}

func (KeyRevokedError) Severity() signal.Severity { return signal.Error }
func (KeyRevokedError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type KeyReadError struct {
	Path string
	Err  error
}

func (e KeyReadError) Error() string {
	return fmt.Sprintf("failed to read key file %q: %v", e.Path, e.Err)
}

func (e KeyReadError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.KeyRead",
		Text: "failed to read SSH-key file {{.Path}}",
		Data: e,
	}
}

func (KeyReadError) Severity() signal.Severity { return signal.Error }
func (KeyReadError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type KeyParseError struct {
	Path        string
	IsPublicKey bool
	Err         error
}

func (e KeyParseError) Error() string {
	return fmt.Sprintf("failed to parse key file %q: %v", e.Path, e.Err)
}

func (e KeyParseError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.KeyParse",
		Text: "failed to parse SSH-key file {{.Path}}",
		Hint: "the provided key-file must contain a valid *private* SSH-key",
		Help: `{{if .IsPublicKey}}found valid *public* SSH-key, while a *private* SSH-key is required{{end}}`,
		Data: e,
	}
}

func (KeyParseError) Severity() signal.Severity { return signal.Error }
func (KeyParseError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type NoAuthMethodError struct{}

func (NoAuthMethodError) Error() string {
	return "no SSH authentication method available (no key specified and SSH agent not available)"
}

func (e NoAuthMethodError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.NoAuthMethod",
		Text: "no SSH authentication method available",
		Hint: "no key specified and SSH agent unavailable",
		Help: "specify a key and/or start SSH agent",
		Data: e,
	}
}

func (NoAuthMethodError) Severity() signal.Severity { return signal.Error }
func (NoAuthMethodError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type AuthError struct {
	Err error
}

func (e AuthError) Error() string {
	return fmt.Sprintf("authentication failed: %v", e.Err)
}

func (e AuthError) Unwrap() error { return e.Err }

func (e AuthError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.Auth",
		Text: "authentication failed: {{.Err}}",
		Data: e,
	}
}

func (AuthError) Severity() signal.Severity { return signal.Error }
func (AuthError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type InvalidTimeoutError struct {
	Value  string
	Source spec.SourceSpan
	Err    error
}

func (e InvalidTimeoutError) Error() string {
	return fmt.Sprintf("invalid timeout %q: %v", e.Value, e.Err)
}

func (e InvalidTimeoutError) Unwrap() error { return e.Err }

func (e InvalidTimeoutError) EventTemplate() event.Template {
	return event.Template{
		ID:     "ssh.InvalidTimeout",
		Text:   `invalid timeout "{{.Value}}"`,
		Hint:   `use a human-readable duration like "2s", "1m30s", or "500ms"`,
		Source: &e.Source,
		Data:   e,
	}
}

func (InvalidTimeoutError) Severity() signal.Severity { return signal.Error }
func (InvalidTimeoutError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type SFTPSessionError struct {
	Err error
}

func (e SFTPSessionError) Error() string {
	return fmt.Sprintf("failed to start SFTP session: %v", e.Err)
}

func (e SFTPSessionError) Unwrap() error { return e.Err }

func (e SFTPSessionError) EventTemplate() event.Template {
	return event.Template{
		ID:   "ssh.SFTPSession",
		Text: "failed to start SFTP session: {{.Err}}",
		Data: e,
	}
}

func (SFTPSessionError) Severity() signal.Severity { return signal.Error }
func (SFTPSessionError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
