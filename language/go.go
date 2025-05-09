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

type Go struct {
	*base
	Root string
}

var goOSArch = &OSArch{
	Linux:  "linux",
	Darwin: "darwin",
	AMD64:  "amd64",
	ARM64:  "arm64",
}

// version, os, arch
const goAssetURL = "https://dl.google.com/go/go%s.%s-%s.tar.gz"

func (g *Go) List(ctx context.Context, all bool) ([]string, error) {
	b, err := HTTPGet(ctx, "https://go.dev/dl/?mode=json&include=all")
	if err != nil {
		return nil, err
	}
	var ress []struct {
		Version string
	}
	if err := json.Unmarshal(b, &ress); err != nil {
		return nil, err
	}
	var out []string
	for _, res := range ress {
		version := strings.TrimPrefix(res.Version, "go")
		out = append(out, version)
	}
	if !all {
		stable := regexp.MustCompile(`^\d+\.\d+(?:\.\d+)$`)
		var out2 []string
		for _, v := range out {
			if stable.MatchString(v) {
				out2 = append(out2, v)
			}
		}
		return out2[:10], nil
	}
	return out, nil
}

func (g *Go) Latest(ctx context.Context) (string, error) {
	out, err := g.List(ctx, false)
	if err != nil {
		return "", err
	}
	return out[0], nil
}

func (g *Go) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		latest, err := g.Latest(ctx)
		if err != nil {
			return "", err
		}
		version = latest
	}
	targetDir := filepath.Join(g.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(goAssetURL, version, goOSArch.OS(), goOSArch.Arch())
	cacheFile := filepath.Join(g.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(g.Root, "cache"), 0o755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return "", err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := g.Untar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}
