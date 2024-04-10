// Package main implements a tool for downloading and organizing security data from Project Discovery's Chaos dataset.
// The tool fetches a list of targets from a JSON index, downloads ZIP archives associated with each target,
// extracts them into dedicated directories, and compiles all text file contents into a single file named 'everything.txt'.
// This file, 'everything.txt', is saved in the directory from which the script is executed, providing a consolidated
// view of the textual data gathered from the downloaded archives.
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// main orchestrates the script's workflow, including creating the base directory, processing the JSON index,
// downloading and extracting data, and finally, compiling the 'everything.txt' file.
func main() {
	// Specify the directory where all operations will take place.
	baseDir := filepath.Join(".", "AllChaosData")

	// Create the base directory.
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create base directory: %v", err)
	}

	// Define the JSON index URL.
	jsonURL := "https://chaos-data.projectdiscovery.io/index.json"

	// Process each entry in the JSON index.
	if err := processURLs(jsonURL, baseDir); err != nil {
		log.Fatalf("Failed to process URLs: %v", err)
	}

	// Compile all .txt files into 'everything.txt', located in the script's execution directory.
	if err := concatenateAllTxtFiles(baseDir, "."); err != nil {
		log.Fatalf("Failed to concatenate all txt files into everything.txt: %v", err)
	}
}

// processURLs fetches the JSON index from the provided URL and processes each entry by downloading
// the associated ZIP file, extracting its contents, and organizing them into directories named after each entry.
func processURLs(jsonURL, baseDir string) error {
	// Fetch the JSON index.
	resp, err := http.Get(jsonURL)
	if err != nil {
		return fmt.Errorf("error fetching JSON index: %w", err)
	}
	defer resp.Body.Close()

	// Decode the JSON index into a slice of entries.
	var entries []struct {
		Name string `json:"name"`
		URL  string `json:"URL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return fmt.Errorf("error decoding JSON index: %w", err)
	}

	// Process each entry in the index.
	for _, entry := range entries {
		fmt.Printf("Processing %s...\n", entry.Name)
		if err := downloadAndUnzip(entry.URL, entry.Name, baseDir); err != nil {
			log.Printf("Failed to process %s: %v\n", entry.Name, err)
		}
	}
	return nil
}

// downloadAndUnzip handles the downloading of a ZIP file from the given URL and extracts its contents
// into a directory within baseDir, named after the entry's "name".
func downloadAndUnzip(url, name, baseDir string) error {
	// Download the ZIP file.
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Create a temporary file to store the ZIP archive.
	tempFile, err := ioutil.TempFile("", "*.zip")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure the temporary file is deleted.

	// Write the downloaded content to the temporary file.
	if _, err = io.Copy(tempFile, resp.Body); err != nil {
		tempFile.Close()
		return fmt.Errorf("error writing to temp file: %w", err)
	}
	tempFile.Close()

	// Create a directory for the entry.
	dirPath := filepath.Join(baseDir, name)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dirPath, err)
	}

	// Extract the ZIP file into the directory.
	if err := unzipFile(tempFile.Name(), dirPath); err != nil {
		return fmt.Errorf("error unzipping file: %w", err)
	}

	return nil
}

// unzipFile extracts the contents of the specified ZIP file into the destination directory.
func unzipFile(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("error opening zip file: %w", err)
	}
	defer r.Close()

	// Iterate through each file in the ZIP archive.
	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		// Create directories if necessary.
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Ensure the parent directory exists.
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Extract the file.
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("error opening output file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("error opening zip content: %w", err)
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("error writing to output file: %w", err)
		}
	}
	return nil
}

// concatenateAllTxtFiles finds all .txt files within baseDir and concatenates their contents
// into a single file named 'everything.txt', placed in the specified output directory.
func concatenateAllTxtFiles(baseDir, outputDir string) error {
	allTxtFiles := findAllTxtFiles(baseDir)

	destPath := filepath.Join(outputDir, "everything.txt")
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", destPath, err)
	}
	defer dest.Close()

	// Concatenate the contents of each .txt file into 'everything.txt'.
	for _, file := range allTxtFiles {
		src, err := os.Open(file)
		if err != nil {
			log.Printf("Failed to open %s for reading: %v", file, err)
			continue
		}

		if _, err = io.Copy(dest, src); err != nil {
			src.Close()
			log.Printf("Failed to copy %s to %s: %v", file, destPath, err)
			continue
		}
		src.Close()

		// Write a newline after each file's content.
		if _, err = dest.WriteString("\n"); err != nil {
			log.Printf("Failed to write newline after %s: %v", file, err)
		}
	}

	fmt.Printf("Successfully created %s with all .txt file content.\n", destPath)
	return nil
}

// findAllTxtFiles recursively finds all .txt files starting from the root directory and returns their paths.
func findAllTxtFiles(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
			files = append(files, path)
		}
		return nil
	})
	return files
}
