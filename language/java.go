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
				if j < 20 {
					out[i*loops+j] = release
				}
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	var out2 []string
	for _, version := range out {
		if version != "" {
			out2 = append(out2, version)
		}
	}
	if all {
		return out2, nil
	}
	majors := map[int]string{}
	for _, version := range out2 {
		m := regexp.MustCompile(`^jdk-?(\d+)`).FindStringSubmatch(version)
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

func (j *Java) Install(ctx context.Context, version string) error {
	if version == "latest" {
		versions, err := j.List(ctx, false)
		if err != nil {
			return err
		}
		version = versions[0]
	}
	targetDir := filepath.Join(j.Root, "versions", version)
	if ExistsFS(targetDir) {
		return errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(javaAssetURL, version, javaOSArch.OS(), javaOSArch.Arch())
	cacheFile := filepath.Join(j.Root, "cache", version+".tar.gz")
	if err := os.MkdirAll(filepath.Join(j.Root, "cache"), 0755); err != nil {
		return err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile); err != nil {
		return err
	}

	if javaOSArch.OS() == "linux" {
		fmt.Println("---> Extracting " + cacheFile)
		if err := Untar(cacheFile, targetDir); err != nil {
			return err
		}
		return nil
	}

	tempTargetDir := filepath.Join(j.Root, "versions", "_"+version)
	defer os.RemoveAll(tempTargetDir)
	fmt.Println("---> Extracting " + cacheFile)
	if err := Untar(cacheFile, tempTargetDir); err != nil {
		return err
	}
	contentsHome := filepath.Join(tempTargetDir, "Contents", "Home")
	if err := os.Rename(contentsHome, targetDir); err != nil {
		return err
	}
	return nil
}

var _ Installer = (*Java)(nil)