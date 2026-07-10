package language

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type GCloud struct {
	*base
	Root string
}

const gcloudComponentsURL = "https://dl.google.com/dl/cloudsdk/channels/rapid/components-2.json"

// version, os, arch
const gcloudAssetURL = "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-%s-%s-%s.tar.gz"

var gcloudOSArch = &OSArch{
	Linux:  "linux",
	Darwin: "darwin",
	AMD64:  "x86_64",
	ARM64:  "arm",
}

var gcloudVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func (g *GCloud) List(ctx context.Context, _ bool) ([]string, error) {
	latest, err := g.Latest(ctx)
	if err != nil {
		return nil, err
	}
	return []string{latest}, nil
}

func (g *GCloud) Latest(ctx context.Context) (string, error) {
	body, err := HTTPGet(ctx, gcloudComponentsURL)
	if err != nil {
		return "", err
	}
	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return "", err
	}
	if !gcloudVersion.MatchString(manifest.Version) {
		return "", errors.New("invalid gcloud version: " + manifest.Version)
	}
	return manifest.Version, nil
}

func (g *GCloud) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		latest, err := g.Latest(ctx)
		if err != nil {
			return "", err
		}
		version = latest
	}
	if !gcloudVersion.MatchString(version) {
		return "", errors.New("invalid version: " + version)
	}

	targetDir := filepath.Join(g.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}
	cacheDir := filepath.Join(g.Root, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}

	gcloudURL := fmt.Sprintf(gcloudAssetURL, version, gcloudOSArch.OS(), gcloudOSArch.Arch())
	gcloudCache := filepath.Join(cacheDir, version+".tar.gz")
	fmt.Println("---> Downloading " + gcloudURL)
	if err := HTTPMirror(ctx, gcloudURL, gcloudCache, nil); err != nil {
		return "", err
	}

	pythonVersion, pythonURL, err := g.latestPython(ctx)
	if err != nil {
		return "", err
	}
	pythonCache := filepath.Join(cacheDir, version+"-python.tar.gz")
	fmt.Printf("---> Downloading Python %s from %s\n", pythonVersion, pythonURL)
	if err := HTTPMirror(ctx, pythonURL, pythonCache, nil); err != nil {
		return "", err
	}

	fmt.Println("---> Extracting " + gcloudCache)
	if err := g.Untar(gcloudCache, targetDir); err != nil {
		return "", err
	}
	return version, nil
}

func (*GCloud) latestPython(ctx context.Context) (string, string, error) {
	p := &Python{}
	versions, err := p.List(ctx, true)
	if err != nil {
		return "", "", err
	}
	for _, version := range versions {
		if !strings.HasPrefix(version, "3.14.") {
			continue
		}
		pythonVersion, tag, ok := strings.Cut(version, "+")
		if !ok {
			return "", "", errors.New("invalid Python version: " + version)
		}
		url := fmt.Sprintf(pythonAssetURL,
			tag, pythonVersion, tag, pythonOSArch.Arch(), pythonOSArch.OS())
		return version, url, nil
	}
	return "", "", errors.New("Python 3.14 not found")
}

func (g *GCloud) Untar(tarball string, targetDir string) error {
	if ExistsFS(targetDir) {
		return errors.New("already exists " + targetDir)
	}
	pythonTarball := strings.TrimSuffix(tarball, ".tar.gz") + "-python.tar.gz"
	if !ExistsFS(pythonTarball) {
		return errors.New("missing Python cache file: " + pythonTarball)
	}

	tempDir := targetDir + "_tmp"
	defer os.RemoveAll(tempDir)
	if err := Untar(tarball, tempDir); err != nil {
		return err
	}
	if err := Untar(pythonTarball, filepath.Join(tempDir, "_python")); err != nil {
		return err
	}
	return os.Rename(tempDir, targetDir)
}

func (g *GCloud) Script(header string, source string, version string, _ string, _ string) string {
	python := filepath.Join(g.Root, "versions", version, "_python", "bin", "python")
	return header + fmt.Sprintf("export CLOUDSDK_PYTHON=\"%s\"\n", python) +
		fmt.Sprintf("exec \"%s\" \"$@\"\n", source)
}
