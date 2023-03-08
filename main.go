package main

import (
	"embed"
	"fmt"
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
	repoName = strings.TrimSuffix(repoName, filepath.Ext(repoName))

	// Capitalize the first letter of the repository name
	capRepoName := strings.Title(repoName)

	// Print the full module path, repository name and capitalized repository name
	fmt.Printf("Full module path: %s\n", modulePath)
	fmt.Printf("Repository name: %s\n", repoName)
	fmt.Printf("Capitalized repository name: %s\n", capRepoName)

	Refactor("nursery", repoName, "*.go", "*.m", "*.md")
	Refactor("Nursery", capRepoName, "*.go", "*.m", "*.md")
}

func Refactor(old, new string, patterns ...string) error {
	return filepath.Walk(".", refactorFunc(old, new, patterns))
}

func refactorFunc(old, new string, filePatterns []string) filepath.WalkFunc {
	return filepath.WalkFunc(func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !!fi.IsDir() {
			return nil
		}

		var matched bool
		for _, pattern := range filePatterns {
			var err error
			matched, err = filepath.Match(pattern, fi.Name())
			if err != nil {
				return err
			}

			if matched {
				read, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				fmt.Println("Refactoring:", path)

				newContents := strings.Replace(string(read), old, new, -1)

				err = os.WriteFile(path, []byte(newContents), 0)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
