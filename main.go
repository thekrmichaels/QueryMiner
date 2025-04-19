package main

import (
	"fmt"
	"github.com/thekrmichaels/QueryMiner/queryminer"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: QueryMiner <source_path> <destination_folder>")
		os.Exit(2)
	}

	sourcePath := os.Args[1]
	destinationFolder := os.Args[2]

	fmt.Printf("Exploring path: %s\n", sourcePath)

	if err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		var walkErr error

		switch {
		case err != nil:
			walkErr = fmt.Errorf("Error accessing to %s: %w", path, err)
		case info.IsDir(), !(strings.HasSuffix(path, ".php") || strings.HasSuffix(path, ".inc")):
			// * Do nothing.
		default:
			fmt.Printf("Scanning file: %s\n", path)

			baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			outputPath := generateUniqueOutputPath(baseName, destinationFolder)

			if createErr := queryminer.GenerateSQLFile(sourcePath, path, outputPath); createErr != nil {
				walkErr = createErr
			}
		}

		return walkErr
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
