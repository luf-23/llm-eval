package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"llm-eval/internal/provider"
)

type Store interface {
	Get(key string) (provider.Response, bool)
	Set(key string, value provider.Response) error
}

type NoopStore struct{}

func (NoopStore) Get(string) (provider.Response, bool) { return provider.Response{}, false }
func (NoopStore) Set(string, provider.Response) error  { return nil }

type FileStore struct {
	dir string
}

func NewFileStore(dir string) FileStore {
	return FileStore{dir: dir}
}

func Key(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (s FileStore) Get(key string) (provider.Response, bool) {
	data, err := os.ReadFile(s.path(key))
	if err != nil {
		return provider.Response{}, false
	}
	var resp provider.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return provider.Response{}, false
	}
	return resp, true
}

func (s FileStore) Set(key string, value provider.Response) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(key), data, 0644)
}

func (s FileStore) path(key string) string {
	return filepath.Join(s.dir, key+".json")
}
