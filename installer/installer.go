package installer

import "context"

type Installer interface {
	List(ctx context.Context, all bool) ([]string, error)
	Install(ctx context.Context, version string) error
}
