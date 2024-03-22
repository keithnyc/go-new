package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the name of the new Go project: ")
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	err := unzip("go-canonical.zip", ".", projectName)
	if err != nil {
		fmt.Println("Error unzipping file:", err)
		return
	}

	err = filepath.Walk(projectName,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Check if the file is a .go file or if it's the go.mod file
			if strings.HasSuffix(info.Name(), ".go") || info.Name() == "go.mod" {
				read, err := os.ReadFile(path) // Changed from ioutil.ReadFile to os.ReadFile

				if err != nil {
					return err
				}

				// Replace "go-canonical" with the new project name
				newContents := strings.Replace(string(read), "go-canonical", projectName, -1)

				// Write the updated contents back to the file using os.WriteFile
				err = os.WriteFile(path, []byte(newContents), info.Mode())
				if err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil {
		fmt.Println("Error replacing text in files:", err)
	}

	fmt.Println("Project initialized successfully.")
}

// unzip extracts a zip archive, renaming the root if necessary.
func unzip(src, dest, projectName string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Replace the root folder name if it is 'go-canonical'
		adjustedPath := strings.TrimPrefix(f.Name, "go-canonical/")
		fpath := filepath.Join(dest, projectName, adjustedPath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without masking the previous error
		closeErr := outFile.Close()
		if err != nil {
			rc.Close()
			return err
		}
		if closeErr != nil {
			rc.Close()
			return closeErr
		}

		rc.Close()
	}

	return nil
}
