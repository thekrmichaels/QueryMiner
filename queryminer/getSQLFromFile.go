package queryminer

import (
	"fmt"
	"os"
	"strings"
)

func captureSQLInQuotes(state *sqlState, lineNumber int, outputFile *os.File) {
	if !state.inString && !state.escaped {
		state.inString = true
		state.stringBuffer.Reset()
		state.sqlStartLine = lineNumber
	} else if state.inString && !state.escaped {
		state.inString = false
		query := state.stringBuffer.String()

		if state.sqlPattern.MatchString(query) {
			state.queryAlreadyCaptured = true

			fmt.Fprintf(outputFile, "-- SQL found in: %s, Line: %d\n", state.relPath, state.sqlStartLine)
			fmt.Fprint(outputFile, normalizeSQL(query, state))
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
	cleanSQL := strings.Replace(strings.Replace(capturedSQL, `$db->query("`, "", -1), `");`, "", -1)

	// * Checks if the query is fragmented (contains PHP variables or dots).
	if strings.ContainsAny(cleanSQL, ".$") {
		fmt.Fprintf(outputFile, "-- ⚠️ SQL possibly fragmented (requires manual review) in: %s, Line: %d\n", state.relPath, state.queryLineStart)
	} else {
		fmt.Fprintf(outputFile, "-- SQL found in: %s, Line: %d\n", state.relPath, state.queryLineStart)
	}

	matches := state.stringInSQLPattern.FindAllStringSubmatch(cleanSQL, -1)

	// * Write valid SQL queries found in the content.
	for _, match := range matches {
		query := match[1]

		if state.sqlPattern.MatchString(query) {
			fmt.Fprint(outputFile, normalizeSQL(query, state))

			return
		}
	}

	fmt.Fprint(outputFile, normalizeSQL(cleanSQL, state))
}

func normalizeSQL(query string, state *sqlState) string {
	// * Pattern to find PHP-style variables like $var or '$var'.
	matches := state.varPattern.FindAllStringSubmatch(query, -1)

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
	replacements := make([]string, 0, len(paramMap)*2)

	for varName, param := range paramMap {
		replacements = append(replacements, "'$"+varName+"'", param)
		replacements = append(replacements, "$"+varName, param)
	}

	replacer := strings.NewReplacer(replacements...)

	return fmt.Sprintf(
		"%s;\n\n",
		strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(replacer.Replace(query)), ";")),
	)
}
