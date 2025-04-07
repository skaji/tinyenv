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
	b, err := HTTPGet(ctx, nodeVersionsURL)
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
		out = append(out, res.Version)
	}
	if all {
		return out, nil
	}
	seen := map[string]string{}
	for _, v := range out {
		major := semver.Major(v)
		if _, ok := seen[major]; !ok {
			seen[major] = v
		}
	}
	out2 := slices.SortedFunc(maps.Values(seen), func(v1, v2 string) int {
		return semver.Compare(v2, v1)
	})
	return out2[:10], nil
}

func (n *Node) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		versions, err := n.List(ctx, false)
		if err != nil {
			return "", err
		}
		version = versions[0]
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
