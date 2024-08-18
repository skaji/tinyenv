package installer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// version, os, arch
const javaAssetURL = "https://api.adoptium.net/v3/binary/latest/%s/ga/%s/%s/jdk/hotspot/normal/eclipse"

func (j *Java) List(_ context.Context, _ bool) ([]string, error) {
	return []string{
		"22", "21", "20", "19", "18", "17", "16", "11", "8",
	}, nil
}

func (j *Java) Install(ctx context.Context, version string) error {
	targetDir := filepath.Join(j.Root, "versions", version)
	if ExistsFS(targetDir) {
		return errors.New("already exists " + targetDir)
	}

	url := fmt.Sprintf(javaAssetURL, version, javaOSArch.OS(), javaOSArch.Arch())
	cacheFile := filepath.Join(j.Root, "cache", fmt.Sprintf("jdk-%s-%s-%s.tar.gz", version, javaOSArch.OS(), javaOSArch.Arch()))
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
