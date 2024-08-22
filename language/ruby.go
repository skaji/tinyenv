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
		versions[i] = "portable-" + tag
	}
	return versions, nil
}

func (r *Ruby) Install(ctx context.Context, version string) error {
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		return errors.New("unsupported os/arch")
	}

	if !strings.HasPrefix(version, "portable-") {
		return errors.New("invalid version: " + version)
	}
	if version == "latest" {
		versions, err := r.List(ctx, true)
		if err != nil {
			return err
		}
		version = versions[0]
	}

	targetDir := filepath.Join(r.Root, "versions", version)
	if ExistsFS(targetDir) {
		return errors.New("already exists " + targetDir)
	}

	url, err := r.url(ctx, version)
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(r.Root, "cache", filepath.Base(url))
	if err := os.MkdirAll(filepath.Join(r.Root, "cache"), 0755); err != nil {
		return err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := UntarStrip(cacheFile, targetDir, 2); err != nil {
		return err
	}
	return nil
}

func (r *Ruby) url(ctx context.Context, version string) (string, error) {
	tag := strings.TrimPrefix(version, "portable-")
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
