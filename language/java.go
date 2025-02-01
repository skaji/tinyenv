package language

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Java struct {
	Root string
}

var javaOSArch = &OSArch{
	Linux:  "linux",
	Darwin: "mac",
	AMD64:  "x64",
	ARM64:  "aarch64",
}

const javaVersionsURL = "https://api.adoptium.net/v3/info/release_names"

// version, os, arch
const javaAssetURL = "https://api.adoptium.net/v3/binary/version/%s/%s/%s/jdk/hotspot/normal/eclipse"

func (j *Java) List(ctx context.Context, all bool) ([]string, error) {
	loops := 5
	out := make([]string, 20*loops)
	var group errgroup.Group
	for i := range loops {
		q := url.Values{}
		q.Set("release_type", "ga")
		q.Set("os", javaOSArch.OS())
		q.Set("architecture", javaOSArch.Arch())
		q.Set("vendor", "eclipse")
		q.Set("project", "jdk")
		q.Set("page_size", "20")
		q.Set("page", strconv.Itoa(i))
		u := javaVersionsURL + "?" + q.Encode()
		group.Go(func() error {
			body, err := HTTPGet(ctx, u)
			if err != nil {
				if strings.HasPrefix(err.Error(), "404") {
					return nil
				}
				return err
			}
			var res struct {
				Releases []string
			}
			if err := json.Unmarshal(body, &res); err != nil {
				return err
			}
			for j, release := range res.Releases {
				out[i*20+j] = release
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	var out2 []string
	for _, version := range out {
		if strings.HasPrefix(version, "jdk-") {
			version = strings.TrimPrefix(version, "jdk-")
			out2 = append(out2, "temurin-"+version)
		}
	}
	if all {
		return out2, nil
	}
	majors := map[int]string{}
	for _, version := range out2 {
		m := regexp.MustCompile(`^temurin-(\d+)`).FindStringSubmatch(version)
		if m != nil {
			if major, err := strconv.Atoi(m[1]); err == nil {
				if _, ok := majors[major]; !ok {
					majors[major] = version
				}
			}
		}
	}
	var out3 []string
	keys := slices.Sorted(maps.Keys(majors))
	slices.Reverse(keys)
	for _, major := range keys {
		out3 = append(out3, majors[major])
	}
	return out3, nil
}

func (j *Java) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		versions, err := j.List(ctx, false)
		if err != nil {
			return "", err
		}
		version = versions[0]
	}
	if !strings.HasPrefix(version, "temurin-") {
		return "", errors.New("invalid version: " + version)
	}

	targetDir := filepath.Join(j.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(javaAssetURL,
		"jdk-"+strings.TrimPrefix(version, "temurin-"),
		javaOSArch.OS(),
		javaOSArch.Arch())
	cacheFile := filepath.Join(j.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(j.Root, "cache"), 0755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return "", err
	}

	fmt.Println("---> Extracting " + cacheFile)
	if err := javaUntar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}

func javaUntar(cacheFile string, targetDir string) error {
	if javaOSArch.OS() == "linux" {
		return Untar(cacheFile, targetDir)
	}

	tempTargetDir := targetDir + "_tmp"
	defer os.RemoveAll(tempTargetDir)
	if err := Untar(cacheFile, tempTargetDir); err != nil {
		return err
	}
	contentsHome := filepath.Join(tempTargetDir, "Contents", "Home")
	return os.Rename(contentsHome, targetDir)
}

func (j *Java) BinDirs() []string {
	return []string{"bin"}
}
