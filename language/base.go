package language

import "regexp"

type base struct{}

func (*base) BinDirs() []string {
	return []string{"bin"}
}

func (*base) BinFile() *regexp.Regexp {
	return nil
}

func (*base) Untar(tarball string, targetDir string) error {
	return Untar(tarball, targetDir)
}
