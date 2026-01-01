package language

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

type Ruby struct {
	*base
	Root string
}

const rubyAPIURL = "https://formulae.brew.sh/api/formula/portable-ruby.json"

func (r *Ruby) list(ctx context.Context) (string, string, error) {
	body, err := HTTPGet(ctx, rubyAPIURL)
	if err != nil {
		return "", "", err
	}
	var res struct {
		Versions struct {
			Stable string
		}
		Bottle struct {
			Stable struct {
				Files map[string]map[string]string
			}
		}
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", "", err
	}
	var find *regexp.Regexp
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			find = regexp.MustCompile(`^catalina$`)
		case "arm64":
			find = regexp.MustCompile(`^arm64_`)
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			find = regexp.MustCompile(`^x86_64_linux$`)
		case "arm64":
			find = regexp.MustCompile(`^arm64_linux$`)
		}
	}
	if find == nil {
		return "", "", fmt.Errorf("unsupported os/arch")
	}
	for key, detail := range res.Bottle.Stable.Files {
		if find.MatchString(key) {
			return "homebrew-portable-" + res.Versions.Stable, detail["url"], nil
		}
	}
	return "", "", fmt.Errorf("cannot find version, url: %v", res.Bottle.Stable.Files)
}

func (r *Ruby) List(ctx context.Context, _ bool) ([]string, error) {
	latest, _, err := r.list(ctx)
	if err != nil {
		return nil, err
	}
	return []string{latest}, nil
}

func (r *Ruby) Latest(ctx context.Context) (string, error) {
	latest, _, err := r.list(ctx)
	if err != nil {
		return "", err
	}
	return latest, nil
}

func (r *Ruby) Install(ctx context.Context, version string) (string, error) {
	latest, url, err := r.list(ctx)
	if err != nil {
		return "", err
	}
	if version == "latest" {
		version = latest
	}
	if version != latest {
		return "", fmt.Errorf("unknown version: %s", version)
	}

	targetDir := filepath.Join(r.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	cacheFile := filepath.Join(r.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(r.Root, "cache"), 0o755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	modifier := func(req *http.Request) {
		req.Header.Add("Authorization", "Bearer QQ==")
	}
	if err := HTTPMirror(ctx, url, cacheFile, modifier); err != nil {
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
