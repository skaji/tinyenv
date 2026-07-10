package language

import (
	"slices"
	"testing"
)

func TestParseAWSVersions(t *testing.T) {
	body := `hash1	refs/tags/2.9.1
hash2	refs/tags/2.10.0
hash3	refs/tags/2.9.2
hash4	refs/tags/1.40.0
hash5	refs/tags/2.10.0dev0
`
	want := []string{"2.10.0", "2.9.2", "2.9.1"}
	a := &AWS{}
	got := a.parseVersions(body, true)
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAWSBinFile(t *testing.T) {
	for _, name := range []string{"aws", "aws_completer"} {
		if !awsBinFile.MatchString(name) {
			t.Errorf("BinFile does not match %q", name)
		}
	}
	for _, name := range []string{"Python", "awscli", "libcrypto.so"} {
		if awsBinFile.MatchString(name) {
			t.Errorf("BinFile unexpectedly matches %q", name)
		}
	}
}
