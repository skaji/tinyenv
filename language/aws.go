package language

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

type AWS struct {
	Root string
}

const awsGitURL = "https://github.com/aws/aws-cli"

var (
	awsBinFile = regexp.MustCompile(`^(?:aws|aws_completer)$`)
	awsVersion = regexp.MustCompile(`^2\.\d+\.\d+$`)
)

func (a *AWS) List(ctx context.Context, all bool) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--refs", awsGitURL, "refs/tags/2*")
	body, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git ls-remote: %w: %s", err, strings.TrimSpace(string(body)))
	}
	return a.parseVersions(string(body), all), nil
}

func (*AWS) parseVersions(body string, all bool) []string {
	var out []string
	for line := range strings.Lines(body) {
		_, tag, ok := strings.Cut(strings.TrimSpace(line), "\trefs/tags/")
		if ok && awsVersion.MatchString(tag) {
			out = append(out, tag)
		}
	}
	slices.SortFunc(out, func(v1, v2 string) int {
		return semver.Compare("v"+v2, "v"+v1)
	})
	if !all && len(out) > 10 {
		out = out[:10]
	}
	return out
}

func (a *AWS) Latest(ctx context.Context) (string, error) {
	versions, err := a.List(ctx, false)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", errors.New("not found")
	}
	return versions[0], nil
}

func (a *AWS) Install(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		latest, err := a.Latest(ctx)
		if err != nil {
			return "", err
		}
		version = latest
	}
	if !awsVersion.MatchString(version) {
		return "", errors.New("invalid version: " + version)
	}

	targetDir := filepath.Join(a.Root, "versions", version)
	if ExistsFS(targetDir) {
		return "", errors.New("already exists " + targetDir)
	}

	url, ext, err := a.asset(version)
	if err != nil {
		return "", err
	}
	cacheFile := filepath.Join(a.Root, "cache", version+ext)
	if err := os.MkdirAll(filepath.Join(a.Root, "cache"), 0o755); err != nil {
		return "", err
	}

	fmt.Println("---> Downloading " + url)
	if err := HTTPMirror(ctx, url, cacheFile, nil); err != nil {
		return "", err
	}
	fmt.Println("---> Extracting " + cacheFile)
	if err := a.Untar(cacheFile, targetDir); err != nil {
		return "", err
	}
	return version, nil
}

func (*AWS) BinDirs() []string {
	return []string{"."}
}

func (*AWS) BinFile() *regexp.Regexp {
	return awsBinFile
}

func (*AWS) Untar(archive string, targetDir string) error {
	if ExistsFS(targetDir) {
		return errors.New("already exists " + targetDir)
	}

	tempDir := targetDir + "_tmp"
	defer os.RemoveAll(tempDir)

	var sourceDir string
	switch runtime.GOOS {
	case "darwin":
		pkgutil, err := exec.LookPath("pkgutil")
		if err != nil {
			return errors.New("missing 'pkgutil' command")
		}
		cmd := exec.Command(pkgutil, "--expand-full", archive, tempDir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		sourceDir = filepath.Join(tempDir, "aws-cli.pkg", "Payload", "aws-cli")
	case "linux":
		unzip, err := exec.LookPath("unzip")
		if err != nil {
			return errors.New("missing 'unzip' command")
		}
		cmd := exec.Command(unzip, "-q", archive, "-d", tempDir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		sourceDir = filepath.Join(tempDir, "aws", "dist")
	default:
		return errors.New("unsupported OS: " + runtime.GOOS)
	}

	if !ExistsFS(sourceDir) {
		return errors.New("missing AWS CLI directory: " + sourceDir)
	}
	return os.Rename(sourceDir, targetDir)
}

func (*AWS) asset(version string) (string, string, error) {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("https://awscli.amazonaws.com/AWSCLIV2-%s.pkg", version), ".pkg", nil
	case "linux":
		var arch string
		switch runtime.GOARCH {
		case "amd64":
			arch = "x86_64"
		case "arm64":
			arch = "aarch64"
		default:
			return "", "", errors.New("unsupported architecture: " + runtime.GOARCH)
		}
		return fmt.Sprintf("https://awscli.amazonaws.com/awscli-exe-linux-%s-%s.zip", arch, version), ".zip", nil
	default:
		return "", "", errors.New("unsupported OS: " + runtime.GOOS)
	}
}
