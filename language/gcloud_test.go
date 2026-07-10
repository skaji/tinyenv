package language

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGCloudScript(t *testing.T) {
	g := &GCloud{Root: "/tinyenv/gcloud"}
	script := g.Script("#!/bin/sh\n# gcloud\n", "/source/gcloud", "575.0.1", "bin", "gcloud")
	python := filepath.Join(g.Root, "versions", "575.0.1", "_python", "bin", "python")
	for _, want := range []string{
		"#!/bin/sh\n# gcloud\n",
		"export CLOUDSDK_PYTHON=\"" + python + "\"\n",
		"exec \"/source/gcloud\" \"$@\"\n",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script does not contain %q:\n%s", want, script)
		}
	}
}
