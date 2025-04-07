package language

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Ruby struct {
	*base
	Root string
}

const rubyURL = "https://github.com/Homebrew/homebrew-portable-ruby"

func (r *Ruby) List(ctx context.Context, all bool) ([]string, error) {
	tags, err := (&GitHub{}).Tags(ctx, rubyURL)
	if err != nil {
		return nil, err
	}
	versions := make([]string, len(tags))
	for i, tag := range tags {
		versions[i] = "homebrew-portable-" + tag
	}
	if !all && len(versions) > 5 {
		return versions[:5], nil
	}
	return versions, nil
}

func (r *Ruby) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		versions, err := r.List(ctx, true)
		if err != nil {
			return "", err
		}
		version = versions[0]
	}
	if !strings.HasPrefix(version, "homebrew-portable-") {
		return "", errors.New("invalid version: " + version)
	}

	targetDir := filepath.Join(r.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url, err := r.url(ctx, version)
	if err != nil {
		return "", err
	}

	cacheFile := filepath.Join(r.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(r.Root, "cache"), 0o755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return "", err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := r.Untar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}

func (r *Ruby) Untar(cacheFile string, targetDir string) error {
	return UntarStrip(cacheFile, targetDir, 2)
}

func (r *Ruby) url(ctx context.Context, version string) (string, error) {
	tag := strings.TrimPrefix(version, "homebrew-portable-")
	assets, err := (&GitHub{}).Assets(ctx, rubyURL, tag)
	if err != nil {
		return "", err
	}
	assetMap := map[string]string{}
	for _, asset := range assets {
		if !strings.HasSuffix(asset, ".tar.gz") {
			continue
		}
		switch {
		case strings.Contains(asset, "x86_64_linux"):
			assetMap["linux_amd64"] = asset
		case strings.Contains(asset, "arm64_linux"):
			assetMap["linux_arm64"] = asset
		case strings.Contains(asset, "arm64"):
			assetMap["darwin_arm64"] = asset
		default:
			assetMap["darwin_amd64"] = asset
		}
	}
	url, ok := assetMap[runtime.GOOS+"_"+runtime.GOARCH]
	if !ok {
		return "", fmt.Errorf("no archive for %s %s", runtime.GOOS, runtime.GOARCH)
	}
	return url, nil
}
