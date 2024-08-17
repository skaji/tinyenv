package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Lang struct {
	Root string
}

func (l *Lang) Version() (string, error) {
	b, err := os.ReadFile(filepath.Join(l.Root, "version"))
	if err != nil {
		return "", err
	}
	return string(b[:len(b)-1]), nil
}

func (l *Lang) SetVersion(version string) error {
	return os.WriteFile(filepath.Join(l.Root, "version"), append([]byte(version), '\n'), 0644)
}

func (l *Lang) Versions() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(l.Root, "versions"))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func (l *Lang) Init() error {
	return os.MkdirAll(filepath.Join(l.Root, "versions"), 0755)
}

func (l *Lang) Rehash() error {
	version, err := l.Version()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(filepath.Join(l.Root, "versions", version, "bin"))
	if err != nil {
		return err
	}
	var exeFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return err
		}
		if info.Mode()&0111 != 0 {
			exeFiles = append(exeFiles, e.Name())
		}
	}
	for _, exeFile := range exeFiles {
		source := filepath.Join(l.Root, "versions", version, "bin", exeFile)
		target := filepath.Join(filepath.Dir(l.Root), "bin", exeFile)
		content := fmt.Sprintf(`#!/bin/sh`+"\n"+`exec "%s" "$@"`+"\n", source)
		if err := os.WriteFile(target, []byte(content), 0755); err != nil {
			return err
		}
	}
	return nil
}
