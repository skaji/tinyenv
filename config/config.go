package config

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

func NewFromFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	var cfg *Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

type Config struct {
	Rehash map[string]*Rehash `json:"rehash"`
}

type Rehash struct {
	Includes []string `json:"includes"`
	Excludes []string `json:"excludes"`
}

func (r *Rehash) Target(name string) bool {
	if r == nil {
		return true
	}
	if len(r.Includes) > 0 {
		if !slices.Contains(r.Includes, name) {
			return false
		}
	}
	if len(r.Excludes) > 0 {
		if slices.Contains(r.Excludes, name) {
			return false
		}
	}
	return true
}
