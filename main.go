package main

import (
	"fmt"
	"github.com/thekrmichaels/QueryMiner/queryminer"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: queryminer <source_path> <destination_folder>")
		os.Exit(2)
	}

	sourcePath := os.Args[1]
	destinationFolder := os.Args[2]

	fmt.Printf("Exploring path: %s\n", sourcePath)

	if err := filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		var walkDirErr error
		ext := filepath.Ext(path)

		switch {
		case err != nil:
			walkDirErr = fmt.Errorf("Error accessing to %s: %w", path, err)
		case d.IsDir(), !(ext == ".php" || ext == ".inc"):
			// * Do nothing.
		default:
			fmt.Printf("Scanning file: %s\n", path)

			baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			outputPath := generateUniqueOutputPath(baseName, destinationFolder)

			if createErr := queryminer.GenerateSQLFile(sourcePath, path, outputPath); createErr != nil {
				walkDirErr = createErr
			}
		}

		return walkDirErr
	}); err != nil {
		fmt.Printf("Error walking through the directory %s: %v\n", sourcePath, err)
		os.Exit(1)
	}
}

func generateUniqueOutputPath(baseName, destinationFolder string) string {
	outputPath := filepath.Join(destinationFolder, baseName+".sql")
	counter := 2

	for {
		if _, statErr := os.Stat(outputPath); os.IsNotExist(statErr) {
			break
		}

		outputPath = filepath.Join(destinationFolder, fmt.Sprintf("%s (%d).sql", baseName, counter))
		counter++
	}

	return outputPath
}
