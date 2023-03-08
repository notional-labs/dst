package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:nursery/*
var nursery embed.FS

func main() {
	// Check if a command-line argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: dst github.com/yourname/yourrepo")
		fmt.Println("yourrepo will be used as the name of the new chain, based on Nursery at https://github.com/notional-labs/nursery")
		return
	}

	// Extract the repository name from the module path
	modulePath := os.Args[1]
	repoName := strings.TrimPrefix(modulePath, "github.com/")
	repoName = strings.Split(repoName, "/")[1]

	// Refactor the contents of the embedded file system
	refactor("nursery", repoName, "*.go", "*.m", "*.md")
	refactor("Nursery", strings.Title(repoName), "*.go", "*.m", "*.md")

	// Write the embedded file system to disk in a folder named after the repository name
	if err := os.Mkdir(repoName, 0o777); err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", repoName, err)
		return
	}
	if err := fs.WalkDir(nursery, ".", fsCopyFunc(repoName, nursery)); err != nil {
		fmt.Printf("Failed to copy embedded filesystem to directory %s: %v\n", repoName, err)
		return
	}
	fmt.Printf("Embedded filesystem written to directory %s\n", repoName)
}

func refactor(old, new string, patterns ...string) error {
	return filepath.Walk(".", func(path string, fi fs.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}

		for _, pattern := range patterns {
			if matched, err := filepath.Match(pattern, fi.Name()); err != nil {
				return err
			} else if matched {
				read, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				fmt.Println("Refactoring:", path)

				newContents := strings.ReplaceAll(string(read), old, new)

				err = os.WriteFile(path, []byte(newContents), 0o666)
				if err != nil {
					return err
				}
				break // move to the next file
			}
		}
		return nil
	})
}

func fsCopyFunc(dst string, src embed.FS) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.Mkdir(filepath.Join(dst, path), 0o666)
		}

		out, err := os.OpenFile(filepath.Join(dst, path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		if err != nil {
			return err
		}
		defer out.Close()

		in, err := src.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return nil
	}
}
