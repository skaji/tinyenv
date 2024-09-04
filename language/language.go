package language

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

var All = []string{
	"go",
	"java",
	"node",
	"perl",
	"python",
	"raku",
	"ruby",
	"solr",
}

type Language struct {
	Name string
	Root string
}

type Specific interface {
	List(ctx context.Context, all bool) ([]string, error)
	Install(ctx context.Context, version string) (string, error)
	BinDirs() []string
}

func (l *Language) Specific() Specific {
	switch l.Name {
	case "go":
		return &Go{Root: l.Root}
	case "java":
		return &Java{Root: l.Root}
	case "node":
		return &Node{Root: l.Root}
	case "perl":
		return &Perl{Root: l.Root}
	case "python":
		return &Python{Root: l.Root}
	case "raku":
		return &Raku{Root: l.Root}
	case "ruby":
		return &Ruby{Root: l.Root}
	case "solr":
		return &Solr{Root: l.Root}
	default:
		panic("unknown language: " + l.Name)
	}
}

func (l *Language) List(ctx context.Context, all bool) ([]string, error) {
	return l.Specific().List(ctx, all)
}

func (l *Language) Install(ctx context.Context, version string) (string, error) {
	return l.Specific().Install(ctx, version)
}

func (l *Language) Version() (string, error) {
	b, err := os.ReadFile(filepath.Join(l.Root, "version"))
	if err != nil {
		return "", errors.New("no version")
	}
	return string(b[:len(b)-1]), nil
}

func (l *Language) SetVersion(version string) error {
	return os.WriteFile(filepath.Join(l.Root, "version"), append([]byte(version), '\n'), 0644)
}

func (l *Language) Versions() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(l.Root, "versions"))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			version := e.Name()
			out = append(out, version)
		}
	}
	slices.SortFunc(out, func(v1, v2 string) int {
		if !strings.HasPrefix(v1, "v") {
			v1 = "v" + v1
		}
		if !strings.HasPrefix(v2, "v") {
			v2 = "v" + v2
		}
		return semver.Compare(v2, v1)
	})
	return out, nil
}

func (l *Language) Init() error {
	versionsDir := filepath.Join(l.Root, "versions")
	if ExistsFS(versionsDir) {
		return nil
	}
	return os.MkdirAll(versionsDir, 0755)
}

func (l *Language) Rehash() error {
	// remove old exeFiles first
	header := fmt.Sprintf("#!/bin/sh\n# %s\n", l.Name)
	headerBytes := []byte(header)
	headerLen := len(headerBytes)
	{
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

	version, err := l.Version()
	if err != nil {
		return err
	}

	for _, binDir := range l.Specific().BinDirs() {
		entries, err := os.ReadDir(filepath.Join(l.Root, "versions", version, binDir))
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
			source := filepath.Join(l.Root, "versions", version, binDir, exeFile)
			target := filepath.Join(filepath.Dir(l.Root), "bin", exeFile)
			content := header + fmt.Sprintf(`exec "%s" "$@"`, source) + "\n"
			if err := os.WriteFile(target, []byte(content), 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Language) Reset(version string) error {
	current, _ := l.Version()
	if version == "-" {
		if current == "" {
			return errors.New("no version set")
		}
		version = current
	}
	targetDir := filepath.Join(l.Root, "versions", version)
	if !ExistsFS(targetDir) {
		return errors.New("invalid version: " + version)
	}
	cacheFile := filepath.Join(l.Root, "cache", version+".tar.gz")
	if !ExistsFS(cacheFile) {
		cacheFile = filepath.Join(l.Root, "cache", version+".tar.xz")
		if !ExistsFS(cacheFile) {
			return errors.New("no cache file for " + version)
		}
	}
	fmt.Println("---> Removing " + targetDir)
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if l.Name == "java" {
		if err := javaUntar(cacheFile, targetDir); err != nil {
			return err
		}
	} else {
		if err := Untar(cacheFile, targetDir); err != nil {
			return err
		}
	}
	if version == current {
		return l.Rehash()
	}
	return nil
}
