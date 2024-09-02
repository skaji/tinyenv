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
	Root string
}

const solrVersionsURL = "https://archive.apache.org/dist/solr/solr/"

// version, version
const solrAssetURL = "https://archive.apache.org/dist/solr/solr/%s/solr-%s.tgz"

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

func (s *Solr) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		versions, err := s.List(ctx, false)
		if err != nil {
			return "", err
		}
		version = versions[0]
	}
	targetDir := filepath.Join(s.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(solrAssetURL, version, version)
	cacheFile := filepath.Join(s.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(s.Root, "cache"), 0755); err != nil {
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
