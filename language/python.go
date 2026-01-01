package language

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

type Python struct {
	*base
	Root string
}

var pythonOSArch = &OSArch{
	Linux:  "unknown-linux-gnu",
	Darwin: "apple-darwin",
	AMD64:  "x86_64",
	ARM64:  "aarch64",
}

const pythonURL = "https://github.com/astral-sh/python-build-standalone"

// tag, pythonVersion, tag, os, arch
const pythonAssetURL = "https://github.com/astral-sh/python-build-standalone/releases/download/%s/cpython-%s+%s-%s-%s-install_only.tar.gz"

func (p *Python) List(ctx context.Context, all bool) ([]string, error) {
	g := &GitHub{}
	tags, err := g.Tags(ctx, pythonURL)
	if err != nil {
		return nil, err
	}
	for i := range 2 {
		latestTag := tags[i]
		assets, err := g.Assets(ctx, pythonURL, latestTag)
		if err != nil {
			return nil, err
		}
		var out []string
		seen := map[string]bool{}
		for _, asset := range assets {
			m := regexp.MustCompile(`cpython-(.+?)\+.*-install_only.tar.gz$`).FindStringSubmatch(asset)
			if m != nil {
				version := m[1] + "+" + latestTag
				if !seen[version] {
					out = append(out, version)
					seen[version] = true
				}
			}
		}
		if len(out) == 0 {
			continue
		}
		slices.SortFunc(out, func(v1, v2 string) int {
			return semver.Compare("v"+v2, "v"+v1)
		})
		if !all && len(out) > 10 {
			out = out[:10]
		}
		return out, nil
	}
	return nil, errors.New("Python.List failed")
}

func (p *Python) Latest(ctx context.Context) (string, error) {
	out, err := p.List(ctx, true)
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", errors.New("not found")
	}
	return out[0], nil
}

func (p *Python) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		latest, err := p.Latest(ctx)
		if err != nil {
			return "", err
		}
		version = latest
	}

	pythonVersion, tag, ok := strings.Cut(version, "+")
	if !ok {
		return "", errors.New("invalid version: " + version)
	}
	targetDir := filepath.Join(p.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(pythonAssetURL,
		tag, pythonVersion, tag, pythonOSArch.Arch(), pythonOSArch.OS())
	cacheFile := filepath.Join(p.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(p.Root, "cache"), 0o755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile, nil); err != nil {
		return "", err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := p.Untar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}
