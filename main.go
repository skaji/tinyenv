package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/skaji/tinyenv/config"
	"github.com/skaji/tinyenv/language"
)

var version = "dev"

var helpMessage = `Usage:
  ❯ tinyenv GLOBAL_COMMAND...
  ❯ tinyenv LANGUAGE COMMAND...

Global Commands:
  %s

Languages:
  %s

Commands:
  %s

Examples:
  ❯ tinyenv versions
  ❯ tinyenv python install -l
  ❯ tinyenv python install 3.9.19+20240814
  ❯ tinyenv python install latest
  ❯ tinyenv python global 3.12.5+20240814
`

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
	globalCommands := []string{
		"files",
		"latest",
		"rehash",
		"root",
		"version",
		"versions",
	}
	languageCommands := []string{
		"global",
		"install",
		"latest",
		"rehash",
		"reset",
		"version",
		"versions",
	}

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "-h", "--help":
			globalCommandStr := strings.Join(globalCommands, "\n  ")
			languageStr := strings.Join(language.All, "\n  ")
			languageCommandStr := strings.Join(languageCommands, "\n  ")
			fmt.Printf(helpMessage, globalCommandStr, languageStr, languageCommandStr)
			os.Exit(1)
		case "--version":
			fmt.Println(version)
			os.Exit(0)
		case "zsh-completions":
			fmt.Print(zshCompletions)
			os.Exit(0)
		case "--completion1":
			for _, l := range language.All {
				fmt.Println(l)
			}
			for _, c := range globalCommands {
				fmt.Println(c)
			}
			os.Exit(0)
		case "--completion2":
			for _, c := range languageCommands {
				fmt.Println(c)
			}
			os.Exit(0)
		}
	}
	if len(os.Args) < 3 &&
		!(len(os.Args) == 2 && slices.Contains(globalCommands, os.Args[1])) {
		fmt.Fprintln(os.Stderr, "invalid arguments")
		os.Exit(1)
	}

	root, err := selectRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var cfg *config.Config
	if path := filepath.Join(root, "config.json"); language.ExistsFS(path) {
		c, err := config.NewFromFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cfg = c
	}

	switch os.Args[1] {
	case "root":
		fmt.Println(root)
		os.Exit(0)
	case "version":
		for _, l := range language.All {
			lang := &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
			if version, err := lang.Version(); err == nil {
				fmt.Printf("%s %s\n", l, version)
			}
		}
		os.Exit(0)
	case "versions":
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
				fmt.Printf("%s%s %s\n", mark, l, v)
			}
		}
		os.Exit(0)
	case "rehash":
		for _, l := range language.All {
			lang := &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
			if err := lang.Rehash(); err != nil {
				fmt.Fprintf(os.Stderr, "%s rehash error: %v", l, err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	case "latest":
		type result struct {
			Language string `json:"language"`
			Latest   string `json:"latest"`
			Have     bool   `json:"have"`
		}
		results := make([]*result, len(language.All))
		var wg sync.WaitGroup
		wg.Add(len(language.All))
		for i, l := range language.All {
			go func() {
				defer wg.Done()
				lang := &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
				latest, err := lang.Latest(context.Background())
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
			}()
		}
		wg.Wait()
		format := "%-5v  %-8s  %s\n"
		fmt.Printf(format, "have?", "language", "latest")
		fmt.Printf(format, "-----", "--------", "------")
		for _, res := range results {
			fmt.Printf(format, res.Have, res.Language, res.Latest)
		}
		os.Exit(0)
	case "files":
		if entries, err := os.ReadDir(filepath.Join(root, "bin")); err == nil {
			for _, e := range entries {
				fmt.Println(filepath.Join(root, "bin", e.Name()))
			}
		}
		entries, err := os.ReadDir(root)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if entry.Name() == "bin" {
				continue
			}
			if !entry.IsDir() {
				continue
			}
			if v := filepath.Join(root, entry.Name(), "version"); language.ExistsFS(v) {
				fmt.Println(v)
			}
			if es, err := os.ReadDir(filepath.Join(root, entry.Name(), "cache")); err == nil {
				for _, e := range es {
					fmt.Println(filepath.Join(root, entry.Name(), "cache", e.Name()))
				}
			}
			if es, err := os.ReadDir(filepath.Join(root, entry.Name(), "versions")); err == nil {
				for _, e := range es {
					fmt.Println(filepath.Join(root, entry.Name(), "versions", e.Name()))
				}
			}
		}
		os.Exit(0)
	}

	var lang *language.Language
	if l := os.Args[1]; slices.Contains(language.All, l) {
		lang = &language.Language{Name: l, Root: filepath.Join(root, l), Config: cfg}
	} else {
		fmt.Fprintln(os.Stderr, "unknown language: "+l)
		os.Exit(1)
	}
	if err := lang.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err2 := func(command string, args ...string) error {
		switch command {
		case "versions":
			vs, err := lang.Versions()
			if err != nil {
				return err
			}
			current, _ := lang.Version()
			bare := len(args) > 0 && args[0] == "--bare"
			for _, v := range vs {
				if bare {
					fmt.Println(v)
					continue
				}
				mark := "  "
				if v == current {
					mark = "* "
				}
				fmt.Println(mark + v)
			}
		case "version":
			v, err := lang.Version()
			if err != nil {
				return err
			}
			fmt.Println(v)
		case "global":
			if len(args) == 0 {
				return errors.New("need version argument")
			}
			version := args[0]
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
		case "latest":
			latest, err := lang.Latest(context.Background())
			if err != nil {
				return err
			}
			fmt.Println(latest)
		case "rehash":
			return lang.Rehash()
		case "install":
			if len(args) == 0 {
				return errors.New("need version argument")
			}
			if args[0] == "-l" || args[0] == "-L" {
				versions, err := lang.List(context.Background(), args[0] == "-L")
				if err != nil {
					return err
				}
				for _, version := range versions {
					fmt.Println(version)
				}
				return nil
			}
			global := false
			if args[0] == "-g" || args[0] == "--global" {
				global = true
				args = args[1:]
			}
			if len(args) == 0 {
				return errors.New("need version argument")
			}
			version := args[0]
			version2, err := lang.Install(context.Background(), version)
			if err != nil || !global {
				return err
			}
			if err := lang.SetVersion(version2); err != nil {
				return err
			}
			return lang.Rehash()
		case "reset":
			if len(args) == 0 {
				return errors.New("need version argument")
			}
			version := args[0]
			return lang.Reset(version)
		default:
			return errors.New("unknown command: " + command)
		}
		return nil
	}(os.Args[2], os.Args[3:]...)
	if err2 != nil {
		fmt.Fprintln(os.Stderr, err2)
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
