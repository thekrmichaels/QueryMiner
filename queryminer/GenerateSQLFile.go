package queryminer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const formattedSQL = "%s;\n\n"

func GenerateSQLFile(sourcePath, path, outputPath string) error {
	var resultErr error
	outputFile, err := os.Create(outputPath)

	if err != nil {
		resultErr = fmt.Errorf("Error creating .sql file %s: %w", outputPath, err)
	}

	defer outputFile.Close()

	file, err := os.Open(path)

	if err != nil {
		resultErr = fmt.Errorf("Error opening file %s: %w", path, err)
	}

	defer file.Close()

	// * Get file location.
	relativePath, err := filepath.Rel(filepath.Dir(sourcePath), path)

	if err != nil {
		relativePath = path
	}

	// * Create a scanner that will read the file line by line.
	fileToScan := bufio.NewScanner(file)

	// * Regular expression that detects classic SQL commands, ignoring upper or lower case ((?i)).
	sqlPattern := regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE)\s+`)

	/*
	 * Stores information about the current state while parsing the code:
	 * - inString: whether we're inside a quoted string.
	 * - escaped: whether the last character was a backslash.
	 * - stringBuffer: accumulates characters of the current SQL string.
	 * - sqlStartLine: records the line where the SQL starts.
	 * - sqlPattern: regex pattern to detect SQL commands.
	 * - relPath: relative path of the file being scanned.
	 */
	state := &sqlState{
		inString:     false,
		escaped:      false,
		stringBuffer: strings.Builder{},
		sqlStartLine: 0,
		sqlPattern:   sqlPattern,
		relPath:      relativePath,
	}

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
	/*
	* Checks if the line contains a call to a SQL query.
	* Reminder: Update the regular expression as more cases are included.
	 */
	matched, _ := regexp.MatchString(`(?i)->(query|exec|prepare)`, line)

	/*
	 * If it is not already within a query call and the line matches the query pattern,
	 * start a new query.
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
