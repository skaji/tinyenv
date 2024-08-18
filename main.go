package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/skaji/tinyenv/installer"
)

var version = "dev"

var helpMessage = `Usage: tinyenv LANGUAGE COMMAND...

Languages:
  go, java, node, perl, python, ruby

Commands:
  global, init, install, reahsh, version, versions

Examples:
  > tinyenv perl init
  > tinyenv python install -l
  > tinyenv python install 3.12.5+20240814
  > tinyenv perl global 5.40.0
  > tinyenv perl version`

//go:embed share/completions.zsh
var zshCompletions string

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println(helpMessage)
		os.Exit(1)
	}
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "zsh-completions" {
		fmt.Print(zshCompletions)
		os.Exit(0)
	}
	if len(os.Args) < 3 && !(len(os.Args) == 2 && os.Args[1] == "root") {
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
	if os.Args[1] == "root" {
		fmt.Println(root)
		os.Exit(0)
	}

	var lang *Lang
	switch l := os.Args[1]; l {
	case "perl", "node", "go", "java", "ruby", "python":
		lang = &Lang{Name: l, Root: filepath.Join(root, l)}
	default:
		fmt.Fprintln(os.Stderr, "unknown language: "+l)
		os.Exit(1)
	}

	err2 := func(command string, args ...string) error {
		switch command {
		case "init":
			if err := lang.Init(); err != nil {
				return err
			}
		case "versions":
			vs, err := lang.Versions()
			if err != nil {
				return err
			}
			current, _ := lang.Version()
			for _, v := range vs {
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
			v := args[0]
			vs, err := lang.Versions()
			if err != nil {
				return err
			}
			if !slices.Contains(vs, v) {
				return errors.New("invalid version: " + v)
			}
			if err := lang.SetVersion(v); err != nil {
				return err
			}
			return lang.Rehash()
		case "rehash":
			return lang.Rehash()
		case "install":
			if len(args) == 0 {
				return errors.New("need version argument.")
			}
			var inst installer.Installer
			switch lang.Name {
			case "go":
				inst = &installer.Go{Root: lang.Root}
			case "java":
				inst = &installer.Java{Root: lang.Root}
			case "node":
				inst = &installer.Node{Root: lang.Root}
			case "perl":
				inst = &installer.Perl{Root: lang.Root}
			case "python":
				inst = &installer.Python{Root: lang.Root}
			default:
				return errors.New("not implemented yet")
			}
			if args[0] == "-l" || args[0] == "-L" {
				versions, err := inst.List(context.Background(), args[0] == "-L")
				if err != nil {
					return err
				}
				for _, version := range versions {
					fmt.Println(version)
				}
				return nil
			}
			version := args[0]
			return inst.Install(context.Background(), version)
		default:
			plugin := "tinyenv-" + command
			path, err := exec.LookPath(plugin)
			if err != nil {
				return errors.New("invalid command: " + command)
			}
			args2 := append([]string{os.Args[1]}, args...)
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
