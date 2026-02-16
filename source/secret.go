// SPDX-License-Identifier: GPL-3.0-only

package source

import (
	"context"

	"godoit.dev/doit/secret"
)

type secretSource struct {
	inner   Source
	backend secret.Backend
}

// WithSecrets wraps src so that LookupSecret delegates to the given backend.
func WithSecrets(src Source, b secret.Backend) Source {
	return &secretSource{inner: src, backend: b}
}

func (s *secretSource) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return s.inner.ReadFile(ctx, path)
}

func (s *secretSource) WriteFile(ctx context.Context, path string, data []byte) error {
	return s.inner.WriteFile(ctx, path, data)
}

func (s *secretSource) EnsureDir(ctx context.Context, path string) error {
	return s.inner.EnsureDir(ctx, path)
}

func (s *secretSource) Stat(ctx context.Context, path string) (FileMeta, error) {
	return s.inner.Stat(ctx, path)
}

func (s *secretSource) LookupEnv(key string) (string, bool) {
	return s.inner.LookupEnv(key)
}

func (s *secretSource) LookupSecret(key string) (string, bool, error) {
	return s.backend.Lookup(key)
}
