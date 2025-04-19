package queryminer

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func writeCapturedSQL(state *sqlState, outputFile *os.File) {
	rawSQL := strings.TrimSpace(state.stringBuffer.String())

	// * Remove matching outer quotes, if present.
	start, end := 0, len(rawSQL)

	if end >= 2 {
		first, last := rawSQL[0], rawSQL[end-1]

		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			start, end = 1, end-1
		}
	}

	query := strings.TrimSpace(rawSQL[start:end])

	if query == "" {
		return
	}

	/*
	* 1. "Flatten" PHP concatenation syntax (e.g. 'SELECT * " . $var . " FROM table' becomes 'SELECT * $var   FROM table')
	* This ensures SQL keywords are clearly visible to the regex pattern without syntax noise.
	 */
	validationQuery := phpConcatPattern.ReplaceAllString(query, " ")

	if !sqlPattern.MatchString(validationQuery) {
		return
	}

	/*
	* 2. Inspect the original query: if the PHP dots are still present, it means the query
	* was physically broken into pieces, flagging it for manual review.
	 */
	if phpConcatPattern.MatchString(query) {
		fmt.Fprintf(outputFile, "-- ⚠️ SQL possibly fragmented (requires manual review) in: %s, Line: %d\n", state.relPath, state.stringStartLine)
	} else {
		fmt.Fprintf(outputFile, "-- SQL found in: %s, Line: %d\n", state.relPath, state.stringStartLine)
	}

	fmt.Fprint(outputFile, normalizeSQL(query))
}

func normalizeSQL(query string) string {
	// * Find all PHP variables inside the SQL statement (e.g. $var or '$var').
	matches := phpVariablesInsideSqlPattern.FindAllStringSubmatch(query, -1)

	// * Assign positional parameters ($1, $2, ...) to each PHP variable ($var).
	paramMap := make(map[string]string)
	paramIndex := 1

	for _, match := range matches {
		varName := match[1]

		// * Only add the variable if it doesn't already exist in the map.
		if _, exists := paramMap[varName]; exists {
			continue
		}

		paramMap[varName] = fmt.Sprintf("$%d", paramIndex)
		paramIndex++
	}

	// * Build every PHP variable representation that can appear in SQL.
	type replacement struct{ old, new string }

	pairs := make([]replacement, 0, len(paramMap)*4) // * 4 representations per variable.

	for varName, param := range paramMap {
		pairs = append(pairs,
			replacement{"'$" + varName + "'", param},
			replacement{`"` + "$" + varName + `"`, param},
			replacement{`\"` + "$" + varName + `\"`, param},
			replacement{"$" + varName, param},
		)
	}

	/*
	 * Replace longer patterns first to prevent shorter ones (e.g., $var)
	 * from matching inside quoted forms (e.g., ‘$var’).
	 */
	sort.Slice(pairs, func(i, j int) bool { return len(pairs[i].old) > len(pairs[j].old) })

	/*
	 * Flatten replacement pairs into the format expected by strings.NewReplacer:
	 * old1, new1, old2, new2, ...
	 */
	replacements := make([]string, 0, len(pairs)*2)

	for _, p := range pairs {
		replacements = append(replacements, p.old, p.new)
	}

	replacer := strings.NewReplacer(replacements...)

	return fmt.Sprintf(
		"%s;\n\n",
		strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(replacer.Replace(query)), ";")),
	)
}
