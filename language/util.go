package language

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

type OSArch struct {
	Linux  string
	Darwin string
	AMD64  string
	ARM64  string
}

func (oa *OSArch) OS() string {
	switch runtime.GOOS {
	case "linux":
		return oa.Linux
	case "darwin":
		return oa.Darwin
	}
	panic("unsupported")
}

func (oa *OSArch) Arch() string {
	switch runtime.GOARCH {
	case "amd64":
		return oa.AMD64
	case "arm64":
		return oa.ARM64
	}
	panic("unsupported")
}

func ExistsFS(target string) bool {
	_, err := os.Stat(target)
	return err == nil
}

func Untar(tarball string, targetDir string) error {
	if _, err := os.Stat(targetDir); err == nil {
		return errors.New("already exists " + targetDir)
	}
	tarExec, err := exec.LookPath("gtar")
	if err != nil {
		tarExec, err = exec.LookPath("tar")
		if err != nil {
			return errors.New("missing 'tar' command")
		}
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	cmd := exec.Command(
		tarExec,
		"xf",
		tarball,
		"-C",
		targetDir,
		"--strip-components=1",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func HTTPGet(ctx context.Context, url string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 != 2 {
		return nil, errors.New(res.Status + " " + url)
	}
	return b, nil
}

func HTTPMirror(ctx context.Context, url string, targetFile string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()
	if res.StatusCode/100 != 2 {
		return errors.New(res.Status + " " + url)
	}

	f, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, res.Body); err != nil {
		return err
	}
	return nil
}
