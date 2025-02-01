package language

type base struct{}

func (*base) BinDirs() []string {
	return []string{"bin"}
}

func (*base) Untar(tarball string, targetDir string) error {
	return Untar(tarball, targetDir)
}
