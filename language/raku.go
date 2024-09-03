package language

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Raku struct {
	Root string
}

var rakuOSArch = &OSArch{
	Linux:  "linux",
	Darwin: "macos",
	AMD64:  "x86_64",
	ARM64:  "arm64",
}

const rakuVersionsURL = "https://rakudo.org/dl/rakudo"

type rakuAsset struct {
	Version string
	URL     string
}

func (r *Raku) list(ctx context.Context) ([]*rakuAsset, error) {
	body, err := HTTPGet(ctx, rakuVersionsURL)
	if err != nil {
		return nil, err
	}
	var ress []struct {
		Type     string
		Platform string
		Arch     string
		Ver      string
		BuildRev int `json:"build_rev"`
		URL      string
	}
	if err := json.Unmarshal(body, &ress); err != nil {
		return nil, err
	}
	var out []*rakuAsset
	for _, res := range ress {
		if res.Type == "archive" && res.Platform == rakuOSArch.OS() && res.Arch == rakuOSArch.Arch() {
			out = append(out, &rakuAsset{
				Version: fmt.Sprintf("%s.%d", res.Ver, res.BuildRev),
				URL:     res.URL,
			})
		}
	}
	slices.SortFunc(out, func(a1, a2 *rakuAsset) int {
		return strings.Compare(a2.Version, a1.Version)
	})
	return out, nil
}

func (r *Raku) List(ctx context.Context, all bool) ([]string, error) {
	assets, err := r.list(ctx)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, asset := range assets {
		out = append(out, asset.Version)
		if !all && len(out) == 10 {
			break
		}
	}
	return out, nil
}

func (r *Raku) Install(ctx context.Context, version string) (string, error) {
	assets, err := r.list(ctx)
	if err != nil {
		return "", err
	}
	var asset *rakuAsset
	if version == "latest" {
		asset = assets[0]
		version = asset.Version
	} else {
		index := slices.IndexFunc(assets, func(a *rakuAsset) bool {
			return a.Version == version
		})
		if index == -1 {
			return "", errors.New("invalid version: " + version)
		}
		asset = assets[index]
	}

	targetDir := filepath.Join(r.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := asset.URL
	cacheFile := filepath.Join(r.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(r.Root, "cache"), 0755); err != nil {
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

func (r *Raku) BinDirs() []string {
	return []string{"bin", filepath.Join("share", "perl6", "site", "bin")}
}
