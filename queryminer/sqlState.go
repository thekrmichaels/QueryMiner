package queryminer

import (
	"regexp"
	"strings"
)

type sqlState struct {
	escaped              bool
	insideQueryCall      bool
	inString             bool
	queryAlreadyCaptured bool
	queryBuffer          strings.Builder
	queryCallPattern     *regexp.Regexp
	queryLineStart       int
	relPath              string
	sqlPattern           *regexp.Regexp
	sqlStartLine         int
	stringBuffer         strings.Builder
	stringInSQLPattern   *regexp.Regexp
	varPattern           *regexp.Regexp
}

/*
newSQLState creates and returns an initial state for parsing SQL queries in a file,
managing string tracking, escaping, text accumulation, location, and SQL pattern detection.
*/
func newSQLState(relPath string) *sqlState {
	return &sqlState{
		queryCallPattern:   regexp.MustCompile(`(?i)(->|::)(query|exec|prepare)`),
		relPath:            relPath,
		sqlPattern:         regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE)\s+`),
		stringInSQLPattern: regexp.MustCompile(`["'](.*?)["']`),
		varPattern:         regexp.MustCompile(`(?:(?:'?\$)(\w+)(?:'?)?)`),
	}
}
