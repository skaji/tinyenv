package installer

import (
	"context"
	"testing"
)

func TestGitHub(t *testing.T) {
	g := &GitHub{}

	tags, err := g.Tags(context.Background(), "https://github.com/indygreg/python-build-standalone")
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range tags {
		t.Log(tag)
	}

	assets, err := g.Assets(context.Background(), "https://github.com/indygreg/python-build-standalone", tags[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, asset := range assets {
		t.Log(asset)
	}
}
