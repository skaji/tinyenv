package language

import "context"

type Installer interface {
	List(ctx context.Context, all bool) ([]string, error)
	Install(ctx context.Context, version string) (string, error)
	BinDirs() []string
}
