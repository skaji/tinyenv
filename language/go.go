package language

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Go struct {
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
	if !all && len(out) > 10 {
		out = out[:10]
	}
	return out, nil
}

func (g *Go) Install(ctx context.Context, version string) error {
	targetDir := filepath.Join(g.Root, "versions", version)
	if ExistsFS(targetDir) {
		return errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(goAssetURL, version, goOSArch.OS(), goOSArch.Arch())
	cacheFile := filepath.Join(g.Root, "cache", filepath.Base(url))
	if err := os.MkdirAll(filepath.Join(g.Root, "cache"), 0755); err != nil {
		return err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := Untar(cacheFile, targetDir); err != nil {
		return err
	}
	return nil
}

var _ Installer = (*Go)(nil)
