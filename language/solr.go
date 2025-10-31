package language

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

type Solr struct {
	*base
	Root string
}

const solrVersionsURL = "https://archive.apache.org/dist/solr/solr/"

// version, version
const solrAssetURL = "https://archive.apache.org/dist/solr/solr/%s/solr-%s.tgz"

// version, version
const solrFastAssetURL = "https://downloads.apache.org/solr/solr/%s/solr-%s.tgz"

func (s *Solr) List(ctx context.Context, all bool) ([]string, error) {
	b, err := HTTPGet(ctx, solrVersionsURL)
	if err != nil {
		return nil, err
	}
	// <a href="9.0.0/">9.0.0/</a>
	matches := regexp.MustCompile(`<a href="([\d.]+)/">`).FindAllStringSubmatch(string(b), -1)
	if len(matches) == 0 {
		return nil, errors.New("no versions for solr")
	}
	var out []string
	for _, m := range matches {
		out = append(out, m[1])
	}
	slices.Reverse(out)
	if !all && len(out) > 10 {
		return out[:10], nil
	}
	return out, nil
}

func (s *Solr) Latest(ctx context.Context) (string, error) {
	out, err := s.List(ctx, true)
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", errors.New("not found")
	}
	return out[0], nil
}

func (s *Solr) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		latest, err := s.Latest(ctx)
		if err != nil {
			return "", err
		}
		version = latest
	}
	targetDir := filepath.Join(s.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(solrFastAssetURL, version, version)
	if err := HTTPHead(ctx, url); err != nil {
		url = fmt.Sprintf(solrAssetURL, version, version)
	}
	cacheFile := filepath.Join(s.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(s.Root, "cache"), 0o755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return "", err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := s.Untar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}
