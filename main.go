package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

var version = "dev"

var helpMessage = `Usage: tinyenv LANGUAGE COMMAND...

Languages:
  go, java, node, perl, python, ruby, rust

Commands:
  global, init, reahsh, version, versions

Examples:
  > tinyenv perl init
  > tinyenv perl global 5.40.0
  > tinyenv perl version
  `

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Print(helpMessage)
		os.Exit(1)
	}
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		os.Exit(0)
	}
	if len(os.Args) < 3 {
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

	var lang *Lang
	switch l := os.Args[1]; l {
	case "perl", "node", "go", "java", "ruby", "python", "rust":
		lang = &Lang{Root: filepath.Join(root, l)}
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
			slices.Sort(vs)
			for _, v := range vs {
				fmt.Println(v)
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
		default:
			return errors.New("invalid command: " + command)
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
