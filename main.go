package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyokomi/emoji/v2"

	git "github.com/go-git/go-git/v5"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	emoji.Println(":rocket: Welcome to the Nurseryd project generator!")
	emoji.Println(":exclamation: Today we will be using the Nursery template.  It features CosmWasm and the Strangelove IBC suite.")

	fmt.Print("Enter the go module path: ")
	modulePath, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("failed to read go module path: %v", err)
	}

	modulePath = strings.TrimSpace(modulePath)
	moduleParts := strings.Split(modulePath, "/")

	if len(moduleParts) < 3 {
		log.Fatalf("invalid go module path: %s", modulePath)
	}

	newName := moduleParts[len(moduleParts)-1]

	// Clone the repository into a folder with the new name
	_, err = git.PlainClone(newName, false, &git.CloneOptions{
		URL:      "https://github.com/notional-labs/nursery",
		Progress: os.Stdout,
	})
	if err != nil {
		log.Fatalf("failed to clone repository: %v", err)
	}

	// Traverse the directory tree and replace the text in files
	err = filepath.Walk(newName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "nurseryd" {
				// Rename the folder with the new name
				newPath := filepath.Join(filepath.Dir(path), newName)
				err := os.Rename(path, newPath)
				if err != nil {
					return err
				}
				return filepath.SkipDir
			}
			return nil
		}
		// Read the contents of the file
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// Replace the text and write back to the file
		newContent := strings.ReplaceAll(string(content), "nursery", newName)
		newContent = strings.ReplaceAll(newContent, "Nursery", strings.Title(newName))
		newContent = strings.ReplaceAll(newContent, "github.com/notional-labs/nursery", modulePath)
		err = ioutil.WriteFile(path, []byte(newContent), info.Mode())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to replace text in files: %v", err)
	}

	// Remove the .git folder
	err = os.RemoveAll(filepath.Join(newName, ".git"))
	if err != nil {
		log.Fatalf("failed to remove .git folder: %v", err)
	}

	emoji.Println(":checkmark: Done!")
	emoji.Println(":question: Next, you can do like: cd " + newName + " && go mod tidy && go install ./...")
	emoji.Println(":mind-blown: Then you can run your new project with: " + newName + "d start")
}
