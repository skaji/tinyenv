package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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
	Includes []*regexp.Regexp
	Excludes []*regexp.Regexp
}

func (r *Rehash) UnmarshalJSON(b []byte) error {
	var data struct {
		Includes []string `json:"includes"`
		Excludes []string `json:"excludes"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	var (
		includes []*regexp.Regexp
		excludes []*regexp.Regexp
	)
	for _, str := range data.Includes {
		reg, err := regexp.Compile(str)
		if err != nil {
			return err
		}
		includes = append(includes, reg)
	}
	for _, str := range data.Excludes {
		reg, err := regexp.Compile(str)
		if err != nil {
			return err
		}
		excludes = append(excludes, reg)
	}
	*r = Rehash{
		Includes: includes,
		Excludes: excludes,
	}
	return nil
}

func (r *Rehash) Target(name string) bool {
	if r == nil {
		return true
	}
	if len(r.Includes) > 0 {
		matched := false
		for _, reg := range r.Includes {
			if reg.MatchString(name) {
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	if len(r.Excludes) > 0 {
		for _, reg := range r.Excludes {
			if reg.MatchString(name) {
				return false
			}
		}
	}
	return true
}
