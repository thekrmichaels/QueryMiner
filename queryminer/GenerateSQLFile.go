package queryminer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func GenerateSQLFile(sourcePath, path, outputPath string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Error creating .sql file %s: %w", outputPath, err)
	}
	defer outputFile.Close()

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Error opening file %s: %w", path, err)
	}
	defer file.Close()

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
		return fmt.Errorf("Error reading file %s: %w", path, err)
	}

	return nil
}

// * Process SQL queries detected in one line of code.
func processLine(line string, lineNumber int, state *sqlState, outputFile *os.File) {
	line, ready := tryInitializeQuery(line, lineNumber, state)

	if !ready {
		return
	}

	for i := range line {
		// * If a closing token is found and handled, stop processing this line early
		if processChar(line[i], state, outputFile) {
			return
		}
	}

	if state.insideQuery {
		state.stringBuffer.WriteByte('\n')
	}
}

// * Handles state changes per character. Returns true if processing for the line should stop.
func processChar(char byte, state *sqlState, outputFile *os.File) bool {
	if char == '\\' {
		state.escaped = true

		state.stringBuffer.WriteByte(char)

		return false
	}

	switch char {
	case '"', '\'':
		handleQuotes(char, state)
	case '(':
		if !state.inString {
			state.parenthesisDepth++
		}
	case ')':
		if !state.inString {
			state.parenthesisDepth--

			if state.captureType == CaptureFunction && state.parenthesisDepth == 0 {
				finalizeQuery(state, outputFile)

				return true
			}
		}
	case ';':
		if !state.inString && state.captureType == CaptureVariable && state.parenthesisDepth == 0 {
			finalizeQuery(state, outputFile)

			return true
		}
	}

	state.stringBuffer.WriteByte(char)

	state.escaped = false

	return false
}

func tryInitializeQuery(line string, lineNumber int, state *sqlState) (string, bool) {
	if state.insideQuery {
		return line, true
	}

	matchFunc := queryCallPattern.FindStringIndex(line)
	matchVar := phpQueryAssignmentStartPattern.FindStringIndex(line)

	if matchFunc != nil {
		initializeCapture(state, lineNumber, CaptureFunction, 1)

		return line[matchFunc[1]:], true
	}

	if matchVar != nil {
		initializeCapture(state, lineNumber, CaptureVariable, 0)

		return line[matchVar[1]-1:], true
	}

	return "", false
}

func initializeCapture(state *sqlState, lineNumber int, captureType CaptureType, parenthesisDepth int) {
	state.captureType = captureType
	state.insideQuery = true
	state.parenthesisDepth = parenthesisDepth

	state.stringBuffer.Reset()

	state.stringStartLine = lineNumber
}

func finalizeQuery(state *sqlState, outputFile *os.File) {
	state.insideQuery = false
	state.captureType = CaptureNone

	writeCapturedSQL(state, outputFile)
}

func handleQuotes(char byte, state *sqlState) {
	if state.escaped {
		return
	}

	if !state.inString {
		state.inString = true
		state.quoteChar = char
	} else if state.quoteChar == char {
		state.inString = false
		state.quoteChar = 0
	}
}
