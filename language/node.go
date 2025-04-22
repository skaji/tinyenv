package language

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"golang.org/x/mod/semver"
)

type Node struct {
	*base
	Root string
}

const nodeVersionsURL = "https://nodejs.org/dist/index.json"

// version, version, os, arch
const nodeAssetURL = "https://nodejs.org/dist/%s/node-%s-%s-%s.tar.xz"

var nodeOSArch = &OSArch{
	Linux:  "linux",
	Darwin: "darwin",
	AMD64:  "x64",
	ARM64:  "arm64",
}

func (n *Node) List(ctx context.Context, all bool) ([]string, error) {
	releases, err := n.list(ctx)
	if err != nil {
		return nil, err
	}
	if all {
		out := make([]string, len(releases))
		for i, r := range releases {
			out[i] = r.Version
		}
		return out, nil
	}

	seen := map[string]string{}
	for _, r := range releases {
		major := semver.Major(r.Version)
		if _, ok := seen[major]; !ok {
			seen[major] = r.Version
		}
	}
	out := slices.SortedFunc(maps.Values(seen), func(v1, v2 string) int {
		return semver.Compare(v2, v1)
	})
	return out[:10], nil
}

type nodeAsset struct {
	Version string          `json:"version"`
	RawLTS  json.RawMessage `json:"lts"`
}

func (r *nodeAsset) LTS() bool {
	var b bool
	if err := json.Unmarshal(r.RawLTS, &b); err == nil {
		return b
	}
	var str string
	if err := json.Unmarshal(r.RawLTS, &str); err == nil {
		return str != ""
	}
	return false
}

func (n *Node) list(ctx context.Context) ([]*nodeAsset, error) {
	b, err := HTTPGet(ctx, nodeVersionsURL)
	if err != nil {
		return nil, err
	}
	var out []*nodeAsset
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	slices.SortFunc(out, func(r1, r2 *nodeAsset) int {
		return semver.Compare(r2.Version, r1.Version)
	})
	return out, nil
}

func (n *Node) Latest(ctx context.Context) (string, error) {
	assets, err := n.list(ctx)
	if err != nil {
		return "", err
	}
	for _, r := range assets {
		if r.LTS() {
			return r.Version, nil
		}
	}
	return "", errors.New("not found")
}

func (n *Node) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		latest, err := n.Latest(ctx)
		if err != nil {
			return "", err
		}
		version = latest
	}
	targetDir := filepath.Join(n.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(nodeAssetURL, version, version, nodeOSArch.OS(), nodeOSArch.Arch())
	cacheFile := filepath.Join(n.Root, "cache", version+".tar.xz")
	if err := os.MkdirAll(filepath.Join(n.Root, "cache"), 0o755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return "", err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := n.Untar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}
