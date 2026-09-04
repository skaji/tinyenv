package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/skaji/tinyenv/config"
	"github.com/skaji/tinyenv/language"
	"github.com/urfave/cli/v3"
)

var version = "dev"

var zshCompletions = `compctl -K _tinyenv tinyenv

_tinyenv() {
  local words completions
  local lang cmd
  read -cA words

  if [[ ${#words} -eq 2 ]]; then
    completions="$(tinyenv --completion1)"
  elif [[ ${#words} -eq 3 ]]; then
    completions="$(tinyenv --completion2)"
  elif [[ ${#words} -eq 4 ]]; then
    lang=$words[2]
    cmd=$words[3]
    if [[ $cmd = global ]]; then
      completions="$(tinyenv $lang versions --bare)"
    fi
  fi
  reply=("${(ps:\n:)completions}")
}
`

func main() {
	globalSpecs := globalCommandsSpec()
	languageSpecs := languageCommandsSpec()

	rootCmd := &cli.Command{
		Name:      "tinyenv",
		Usage:     "A tiny replacement of *env (rbenv, plenv, goenv, ...)",
		UsageText: "tinyenv GLOBAL_COMMAND...\n  tinyenv LANGUAGE COMMAND...",
		Version:   version,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "completion1", Hidden: true},
			&cli.BoolFlag{Name: "completion2", Hidden: true},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			if cmd.Bool("completion1") {
				for _, l := range language.All {
					fmt.Fprintln(cmd.Writer, l)
				}
				for _, c := range globalSpecs {
					fmt.Fprintln(cmd.Writer, c.Name)
				}
				return nil
			}
			if cmd.Bool("completion2") {
				if len(languageSpecs) > 0 {
					for _, c := range languageSpecs[0].Commands {
						fmt.Fprintln(cmd.Writer, c.Name)
					}
				}
				return nil
			}
			return cli.Exit("invalid arguments", 1)
		},
		CommandNotFound: func(_ context.Context, cmd *cli.Command, name string) {
			fmt.Fprintln(cmd.ErrWriter, "unknown language: "+name)
			cli.OsExiter(1)
		},
		Commands: append(globalSpecs, languageSpecs...),
	}
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintln(cmd.Writer, cmd.Version)
	}

	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func selectRoot() (string, error) {
	root := os.Getenv("TINYENV_ROOT")
	if root == "" {
		executable, err := os.Executable()
		if err != nil {
			return "", err
		}
		root = filepath.Dir(filepath.Dir(executable))
	}
	return filepath.Abs(root)
}

func globalCommandsSpec() []*cli.Command {
	return []*cli.Command{
		{
			Name: "root",
			Action: func(_ context.Context, cmd *cli.Command) error {
				root, _, err := loadRootConfig()
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.Writer, root)
				return nil
			},
		},
		{
			Name: "version",
			Action: func(_ context.Context, cmd *cli.Command) error {
				root, cfg, err := loadRootConfig()
				if err != nil {
					return err
				}
				for _, l := range language.All {
					lang := &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
					if version, err := lang.Version(); err == nil {
						fmt.Fprintf(cmd.Writer, "%s %s\n", l, version)
					}
				}
				return nil
			},
		},
		{
			Name: "versions",
			Action: func(_ context.Context, cmd *cli.Command) error {
				root, cfg, err := loadRootConfig()
				if err != nil {
					return err
				}
				for _, l := range language.All {
					lang := &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
					versions, err := lang.Versions()
					if err != nil {
						continue
					}
					version, _ := lang.Version()
					for _, v := range versions {
						mark := "  "
						if v == version {
							mark = "* "
						}
						fmt.Fprintf(cmd.Writer, "%s%s %s\n", mark, l, v)
					}
				}
				return nil
			},
		},
		{
			Name: "rehash",
			Action: func(_ context.Context, _ *cli.Command) error {
				root, cfg, err := loadRootConfig()
				if err != nil {
					return err
				}
				for _, l := range language.All {
					lang := &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
					if err := lang.Rehash(); err != nil {
						return cli.Exit(fmt.Sprintf("%s rehash error: %v", l, err), 1)
					}
				}
				return nil
			},
		},
		{
			Name: "latest",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				root, cfg, err := loadRootConfig()
				if err != nil {
					return err
				}
				type result struct {
					Language string `json:"language"`
					Latest   string `json:"latest"`
					Have     bool   `json:"have"`
				}
				results := make([]*result, len(language.All))
				var wg sync.WaitGroup
				wg.Add(len(language.All))
				for i, l := range language.All {
					go func(i int, l string) {
						defer wg.Done()
						lang := &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
						latest, err := lang.Latest(ctx)
						if err != nil {
							results[i] = &result{
								Language: l,
								Latest:   "error: " + err.Error(),
								Have:     false,
							}
							return
						}
						locals, _ := lang.Versions()
						have := slices.Contains(locals, latest)
						results[i] = &result{
							Language: l,
							Latest:   latest,
							Have:     have,
						}
					}(i, l)
				}
				wg.Wait()
				format := "%-5v  %-8s  %s\n"
				fmt.Fprintf(cmd.Writer, format, "have?", "language", "latest")
				fmt.Fprintf(cmd.Writer, format, "-----", "--------", "------")
				for _, res := range results {
					fmt.Fprintf(cmd.Writer, format, res.Have, res.Language, res.Latest)
				}
				return nil
			},
		},
		{
			Name: "files",
			Action: func(_ context.Context, cmd *cli.Command) error {
				root, _, err := loadRootConfig()
				if err != nil {
					return err
				}
				if entries, err := os.ReadDir(filepath.Join(root, "bin")); err == nil {
					for _, e := range entries {
						fmt.Fprintln(cmd.Writer, filepath.Join(root, "bin", e.Name()))
					}
				}
				entries, err := os.ReadDir(root)
				if err != nil {
					return err
				}
				for _, entry := range entries {
					if entry.Name() == "bin" {
						continue
					}
					if !entry.IsDir() {
						continue
					}
					if v := filepath.Join(root, entry.Name(), "version"); language.ExistsFS(v) {
						fmt.Fprintln(cmd.Writer, v)
					}
					if es, err := os.ReadDir(filepath.Join(root, entry.Name(), "cache")); err == nil {
						for _, e := range es {
							fmt.Fprintln(cmd.Writer, filepath.Join(root, entry.Name(), "cache", e.Name()))
						}
					}
					if es, err := os.ReadDir(filepath.Join(root, entry.Name(), "versions")); err == nil {
						for _, e := range es {
							fmt.Fprintln(cmd.Writer, filepath.Join(root, entry.Name(), "versions", e.Name()))
						}
					}
				}
				return nil
			},
		},
		{
			Name: "zsh-completions",
			Action: func(_ context.Context, cmd *cli.Command) error {
				fmt.Fprint(cmd.Writer, zshCompletions)
				return nil
			},
		},
	}
}

func languageCommandsSpec() []*cli.Command {
	commands := make([]*cli.Command, 0, len(language.All))
	for _, name := range language.All {
		langName := name
		commands = append(commands, &cli.Command{
			Name:  langName,
			Usage: langName + " commands",
			Action: func(_ context.Context, _ *cli.Command) error {
				return cli.Exit("invalid arguments", 1)
			},
			Commands: languageSubcommandsSpec(langName),
		})
	}
	return commands
}

func languageSubcommandsSpec(langName string) []*cli.Command {
	return []*cli.Command{
		{
			Name: "versions",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "bare"},
			},
			Action: func(_ context.Context, cmd *cli.Command) error {
				lang, err := loadLanguageConfig(langName)
				if err != nil {
					return err
				}
				vs, err := lang.Versions()
				if err != nil {
					return err
				}
				current, _ := lang.Version()
				bare := cmd.Bool("bare")
				for _, v := range vs {
					if bare {
						fmt.Fprintln(cmd.Writer, v)
						continue
					}
					mark := "  "
					if v == current {
						mark = "* "
					}
					fmt.Fprintln(cmd.Writer, mark+v)
				}
				return nil
			},
		},
		{
			Name: "version",
			Action: func(_ context.Context, cmd *cli.Command) error {
				lang, err := loadLanguageConfig(langName)
				if err != nil {
					return err
				}
				v, err := lang.Version()
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.Writer, v)
				return nil
			},
		},
		{
			Name: "global",
			Action: func(_ context.Context, cmd *cli.Command) error {
				lang, err := loadLanguageConfig(langName)
				if err != nil {
					return err
				}
				if cmd.Args().Len() == 0 {
					return errors.New("need version argument")
				}
				version := cmd.Args().First()
				versions, err := lang.Versions()
				if err != nil {
					return err
				}
				if !slices.Contains(versions, version) {
					return errors.New("invalid version: " + version)
				}
				if err := lang.SetVersion(version); err != nil {
					return err
				}
				return lang.Rehash()
			},
		},
		{
			Name: "latest",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				lang, err := loadLanguageConfig(langName)
				if err != nil {
					return err
				}
				latest, err := lang.Latest(ctx)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.Writer, latest)
				return nil
			},
		},
		{
			Name: "rehash",
			Action: func(_ context.Context, _ *cli.Command) error {
				lang, err := loadLanguageConfig(langName)
				if err != nil {
					return err
				}
				return lang.Rehash()
			},
		},
		{
			Name: "install",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "list", Aliases: []string{"l"}},
				&cli.BoolFlag{Name: "list-all", Aliases: []string{"L"}},
				&cli.BoolFlag{Name: "global", Aliases: []string{"g"}},
				&cli.StringFlag{Name: "target", Aliases: []string{"t"}},
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				lang, err := loadLanguageConfig(langName)
				if err != nil {
					return err
				}
				list := cmd.Bool("list") || cmd.Bool("list-all")
				if list {
					versions, err := lang.List(ctx, cmd.Bool("list-all"))
					if err != nil {
						return err
					}
					for _, version := range versions {
						fmt.Fprintln(cmd.Writer, version)
					}
					return nil
				}
				if cmd.Args().Len() == 0 {
					return errors.New("need version argument")
				}
				targetDir := cmd.String("target")
				version := cmd.Args().First()
				version2, err := lang.Install(ctx, version, targetDir)
				if err != nil {
					return err
				}
				if targetDir != "" {
					return nil
				}
				if !cmd.Bool("global") {
					return nil
				}
				if err := lang.SetVersion(version2); err != nil {
					return err
				}
				return lang.Rehash()
			},
		},
		{
			Name: "reset",
			Action: func(_ context.Context, cmd *cli.Command) error {
				lang, err := loadLanguageConfig(langName)
				if err != nil {
					return err
				}
				if cmd.Args().Len() == 0 {
					return errors.New("need version argument")
				}
				version := cmd.Args().First()
				return lang.Reset(version)
			},
		},
	}
}

func loadRootConfig() (string, *config.Config, error) {
	root, err := selectRoot()
	if err != nil {
		return "", nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		return "", nil, err
	}
	var cfg *config.Config
	if path := filepath.Join(root, "config.json"); language.ExistsFS(path) {
		c, err := config.NewFromFile(path)
		if err != nil {
			return "", nil, err
		}
		cfg = c
	}
	return root, cfg, nil
}

func loadLanguageConfig(name string) (*language.Language, error) {
	root, cfg, err := loadRootConfig()
	if err != nil {
		return nil, err
	}
	lang := &language.Language{Name: name, Root: filepath.Join(root, name), Config: cfg}
	if err := lang.Init(); err != nil {
		return nil, err
	}
	return lang, nil
}
