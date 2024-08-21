package language

import (
	"context"
	"testing"
)

func TestPythonList(t *testing.T) {
	p := &Python{}
	versions, err := p.List(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(versions)
}
