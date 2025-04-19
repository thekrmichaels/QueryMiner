package queryminer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerateSQLFile(sourcePath, path, outputPath string) error {
	var resultErr error
	outputFile, err := os.Create(outputPath)

	if err != nil {
		resultErr = fmt.Errorf("Error creating .sql file %s: %w", outputPath, err)
	} else {
		defer outputFile.Close()
	}

	file, err := os.Open(path)

	if err != nil {
		resultErr = fmt.Errorf("Error opening file %s: %w", path, err)
	} else {
		defer file.Close()
	}

	// * Get file location.
	relativePath, err := filepath.Rel(filepath.Dir(sourcePath), path)

	if err != nil {
		relativePath = path
	}

	fileToScan := bufio.NewScanner(file)
	state := newSQLState(relativePath)

	fmt.Fprintf(outputFile, "-- SQL Queries extracted from: %s\n\n", relativePath)

	for lineNumber := 1; fileToScan.Scan(); lineNumber++ {
		line := fileToScan.Text()

		processLine(line, lineNumber, state, outputFile)
	}

	if err := fileToScan.Err(); err != nil {
		resultErr = fmt.Errorf("Error reading file %s: %w", path, err)
	}

	return resultErr
}

// * Processes SQL query calls detected in one line of code.
func processLine(line string, lineNumber int, state *sqlState, outputFile *os.File) {
	matched := state.queryCallPattern.MatchString(line)

	/*
	 * If it isn't already in a query call, but the line matches the query pattern,
	 * start a new one.
	 */
	if !state.insideQueryCall && matched {
		state.insideQueryCall = true
		state.queryBuffer.Reset()
		state.queryLineStart = lineNumber
	}

	for i := range len(line) {
		char := line[i : i+1]

		// * Handle special characters: escaped characters, quotes, or append to string buffer.
		switch char {
		case "\\":
			state.escaped = !state.escaped
		case "\"":
			captureSQLInQuotes(state, lineNumber, outputFile)
		default:
			if state.inString {
				state.stringBuffer.WriteString(char)
			}

			if char != "\\" {
				state.escaped = false
			}
		}
	}

	if state.insideQueryCall {
		state.queryBuffer.WriteString(line + "\n")

		// * Checks if the end of the query has been found.
		if strings.Contains(line, ");") {
			state.insideQueryCall = false
			writeCapturedSQL(state, outputFile)
		}
	}
}
