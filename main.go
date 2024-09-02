package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/skaji/tinyenv/language"
)

var version = "dev"

var helpMessage = `Usage:
  ❯ tinyenv GLOBAL_COMMAND...
  ❯ tinyenv LANGUAGE COMMAND...

Global Commands:
  version, versions

Languages:
  go, java, node, perl, python, raku, ruby
  solr

Commands:
  global, install, reahsh, reset, version, versions

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
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "-h", "--help":
			fmt.Print(helpMessage)
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
			fmt.Println("version")
			fmt.Println("versions")
			os.Exit(0)
		case "--completion2":
			fmt.Println("global")
			fmt.Println("install")
			fmt.Println("rehash")
			fmt.Println("reset")
			fmt.Println("version")
			fmt.Println("versions")
			os.Exit(0)
		}
	}
	if len(os.Args) < 3 && !(len(os.Args) == 2 && (os.Args[1] == "root" || os.Args[1] == "versions" || os.Args[1] == "version")) {
		fmt.Fprintln(os.Stderr, "invalid arguments")
		os.Exit(1)
	}

	root, err := selectRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch os.Args[1] {
	case "root":
		fmt.Println(root)
		os.Exit(0)
	case "version":
		for _, l := range language.All {
			lang := &language.Language{Name: l, Root: filepath.Join(root, l)}
			if version, err := lang.Version(); err == nil {
				fmt.Printf("%s %s\n", l, version)
			}
		}
		os.Exit(0)
	case "versions":
		for _, l := range language.All {
			lang := &language.Language{Name: l, Root: filepath.Join(root, l)}
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
	}

	var lang *language.Language
	if l := os.Args[1]; slices.Contains(language.All, l) {
		lang = &language.Language{Name: l, Root: filepath.Join(root, l)}
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
				return errors.New("need version argument.")
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
		case "rehash":
			return lang.Rehash()
		case "install":
			installer := lang.Installer()
			if installer == nil {
				return errors.New("no installer for " + lang.Name)
			}
			if len(args) == 0 {
				return errors.New("need version argument.")
			}
			if args[0] == "-l" || args[0] == "-L" {
				versions, err := installer.List(context.Background(), args[0] == "-L")
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
				return errors.New("need version argument.")
			}
			version := args[0]
			version2, err := installer.Install(context.Background(), version)
			if err != nil || !global {
				return err
			}
			if err := lang.SetVersion(version2); err != nil {
				return err
			}
			return lang.Rehash()
		case "reset":
			if len(args) == 0 {
				return errors.New("need version argument.")
			}
			version := args[0]
			return lang.Reset(version)
		default:
			plugin := "tinyenv-" + command
			path, err := exec.LookPath(plugin)
			if err != nil {
				return errors.New("invalid command: " + command)
			}
			args2 := append([]string{os.Args[1], command}, args...)
			cmd := exec.Command(path, args2...)
			cmd.Env = append(slices.Clone(os.Environ()), "TINYENV_ROOT="+root)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
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
