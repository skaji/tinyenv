package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

type Lang struct {
	Name string
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

	header := fmt.Sprintf("#!/bin/sh\n# %s\n", l.Name)
	headerBytes := []byte(header)
	headerLen := len(headerBytes)
	{
		// remove old exeFiles
		rootBinDir := filepath.Join(filepath.Dir(l.Root), "bin")
		entries, err := os.ReadDir(rootBinDir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			path := filepath.Join(rootBinDir, e.Name())
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			b := make([]byte, headerLen)
			_, err = f.Read(b)
			f.Close()
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			if slices.Equal(b[:headerLen], headerBytes) {
				if err := os.Remove(path); err != nil {
					return err
				}
			}
		}
	}

	for _, exeFile := range exeFiles {
		source := filepath.Join(l.Root, "versions", version, "bin", exeFile)
		target := filepath.Join(filepath.Dir(l.Root), "bin", exeFile)
		content := header + fmt.Sprintf(`exec "%s" "$@"`, source) + "\n"
		if err := os.WriteFile(target, []byte(content), 0755); err != nil {
			return err
		}
	}
	return nil
}
