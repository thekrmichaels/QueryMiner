package queryminer

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func captureSQLInQuotes(state *sqlState, lineNumber int, outputFile *os.File) {
	if !state.inString && !state.escaped {
		state.inString = true
		state.stringBuffer.Reset()
		state.sqlStartLine = lineNumber
	} else if state.inString && !state.escaped {
		state.inString = false

		if state.sqlPattern.MatchString(state.stringBuffer.String()) {
			state.queryAlreadyCaptured = true
			fmt.Fprintf(outputFile, "-- SQL found in: %s, Line: %d\n", state.relPath, state.sqlStartLine)
			fmt.Fprintf(outputFile, formattedSQL, normalizeSQL(state.stringBuffer.String()))
		}
	}
}

func writeCapturedSQL(state *sqlState, outputFile *os.File) {
	if state.queryAlreadyCaptured {
		state.queryAlreadyCaptured = false
		return
	}

	capturedSQL := state.queryBuffer.String()

	/*
	* Remove the syntax $db->query(“”); to isolate the SQL statement.
	* Reminder: Update the regular expression as more cases are included.
	 */
	cleanSQL := strings.Replace(capturedSQL, `$db->query("`, "", -1)
	cleanSQL = strings.Replace(cleanSQL, `");`, "", -1)

	// * Checks if the query is fragmented (contains PHP variables or dots).
	if strings.Contains(cleanSQL, ".") || strings.Contains(cleanSQL, "$") {
		fmt.Fprintf(outputFile, "-- ⚠️ SQL possibly fragmented (requires manual review) in: %s, Line: %d\n", state.relPath, state.queryLineStart)
	} else {
		fmt.Fprintf(outputFile, "-- SQL found in: %s, Line: %d\n", state.relPath, state.queryLineStart)
	}

	stringInSQLPattern := regexp.MustCompile(`["'](.*?)["']`)
	matches := stringInSQLPattern.FindAllStringSubmatch(cleanSQL, -1)

	// * Write valid SQL queries found in the content.
	for _, match := range matches {
		sql := match[1]
		if state.sqlPattern.MatchString(sql) {
			fmt.Fprintf(outputFile, formattedSQL, normalizeSQL(sql))
			return
		}
	}

	fmt.Fprintf(outputFile, formattedSQL, normalizeSQL(cleanSQL))
}

func normalizeSQL(query string) string {
	// * Pattern to find PHP-style variables like $var or '$var'.
	varPattern := regexp.MustCompile(`(?:(?:'?\$)(\w+)(?:'?)?)`)
	matches := varPattern.FindAllStringSubmatch(query, -1)

	/*
	* Assigns positional parameters ($1, $2, ...) to each PHP variable ($var),
	* avoiding duplicates and substituting both forms: '$var' and $var.
	 */
	paramMap := make(map[string]string)

	for paramIndex, match := range matches {
		varName := match[1]
		if _, exists := paramMap[varName]; !exists {
			paramMap[varName] = fmt.Sprintf("$%d", paramIndex+1)
		}
	}

	// * Replace all occurrences of the variable by its positional parameter.
	for varName, param := range paramMap {
		query = strings.ReplaceAll(query, "'$"+varName+"'", param)
		query = strings.ReplaceAll(query, "$"+varName, param)
	}

	// * Remove redundant semicolon (if any).
	return strings.TrimSuffix(query, ";")
}
