package language

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Perl struct {
	Root string
}

var perlOSArch = &OSArch{
	Linux:  "linux",
	Darwin: "darwin",
	AMD64:  "amd64",
	ARM64:  "arm64",
}

const perlVersionsURL = "https://raw.githubusercontent.com/skaji/relocatable-perl/main/releases.csv"

// version, os, arch
const perlAssetURL = "https://github.com/skaji/relocatable-perl/releases/download/%s/perl-%s-%s.tar.xz"

func (p *Perl) List(ctx context.Context, all bool) ([]string, error) {
	b, err := HTTPGet(ctx, perlVersionsURL)
	if err != nil {
		return nil, err
	}
	var out []string
	seen := map[string]bool{}
	for i, line := range strings.Split(string(b), "\n") {
		if i == 0 {
			continue
		}
		if !strings.Contains(line, ","+perlOSArch.OS()+",") {
			continue
		}
		if !strings.Contains(line, ","+perlOSArch.Arch()+",") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) > 0 {
			version := parts[0]
			if !seen[version] {
				out = append(out, "relocatable-"+version)
				seen[version] = true
			}
		}
	}
	if !all && len(out) > 10 {
		out = out[:10]
	}
	return out, nil
}

func (p *Perl) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		versions, err := p.List(ctx, false)
		if err != nil {
			return "", err
		}
		version = versions[0]
	}
	if !strings.HasPrefix(version, "relocatable-") {
		return "", errors.New("invalid version")
	}
	targetDir := filepath.Join(p.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(perlAssetURL, strings.TrimPrefix(version, "relocatable-"), perlOSArch.OS(), perlOSArch.Arch())
	cacheFile := filepath.Join(p.Root, "cache", version+".tar.xz")
	if err := os.MkdirAll(filepath.Join(p.Root, "cache"), 0755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return "", err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := Untar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}

var _ Installer = (*Perl)(nil)
