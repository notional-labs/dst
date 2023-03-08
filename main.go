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

	// Get the full module path from the command-line argument
	modulePath := os.Args[1]

	// Extract the repository name from the module path
	repoName := strings.TrimPrefix(modulePath, "github.com/")
	repoName = strings.Split(repoName, "/")[1]

	// Print the full module path and repository name
	fmt.Printf("Full module path: %s\n", modulePath)
	fmt.Printf("Repository name: %s\n", repoName)

	// Refactor the contents of the embedded file system
	err := Refactor(nursery, repoName, "*.go", "*.m", "*.md")
	if err != nil {
		fmt.Printf("Failed to refactor embedded filesystem: %v\n", err)
		return
	}

	// Write the embedded file system to disk in a folder named after the repository name
	if err := os.Mkdir(repoName, 0o777); err != nil && !os.IsExist(err) {
		fmt.Printf("Failed to create directory %s: %v\n", repoName, err)
		return
	}
	if err := fsCopy(repoName, nursery); err != nil {
		fmt.Printf("Failed to copy embedded filesystem to directory %s: %v\n", repoName, err)
		return
	}
	fmt.Printf("Embedded filesystem written to directory %s\n", repoName)
}

// Refactor the contents of an embedded file system
func Refactor(embedFS embed.FS, new string, patterns ...string) error {
	return fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(path, "./.") {
			return nil
		}

		var matched bool
		for _, pattern := range patterns {
			matched, err = filepath.Match(pattern, d.Name())
			if err != nil {
				return err
			}

			if matched {
				read, err := embedFS.ReadFile(path)
				if err != nil {
					return err
				}

				fmt.Println("Refactoring:", path)

				newContents := strings.Replace(string(read), "Nursery", new, -1)
				newContents = strings.Replace(newContents, "nursery", strings.ToLower(new), -1)

				err = os.WriteFile(path, []byte(newContents), 0)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// Copy the embedded filesystem to disk
func fsCopy(dst string, src embed.FS) error {
	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(dst, 0o777); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dst, err)
	}

	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Create the destination directory if it doesn't exist
			subdir := filepath.Join(dst, path)
			if err := os.MkdirAll(subdir, d.Type()); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", subdir, err)
			}
			return nil
		}

		// Rename any file named go.m to go.mod
		if filepath.Base(path) == "go.m" {
			path = filepath.Dir(path) + "/go.mod"
		}

		// Copy the file to the destination
		in, err := src.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.OpenFile(filepath.Join(dst, path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o777)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err = io.Copy(out, in); err != nil {
			return err
		}

		return nil
	})
}
