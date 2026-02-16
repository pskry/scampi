// SPDX-License-Identifier: GPL-3.0-only

package secret

import (
	"encoding/json"
	"fmt"
)

// FileBackend loads secrets from a flat JSON key-value file.
type FileBackend struct {
	secrets map[string]string
}

func NewFileBackend(data []byte) (*FileBackend, error) {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing secrets file: %w", err)
	}
	return &FileBackend{secrets: raw}, nil
}

func (f *FileBackend) Name() string { return "unencrypted_file" }

func (f *FileBackend) Lookup(key string) (string, bool, error) {
	v, ok := f.secrets[key]
	return v, ok, nil
}
