package language

import "fmt"

type base struct{}

func (*base) BinDirs() []string {
	return []string{"bin"}
}

func (*base) Untar(tarball string, targetDir string) error {
	return Untar(tarball, targetDir)
}

func (*base) Script(header string, source string, _ string, _ string, _ string) string {
	return header + fmt.Sprintf(`exec "%s" "$@"`, source) + "\n"
}
