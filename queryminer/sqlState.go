package queryminer

import (
	"regexp"
	"strings"
)

type sqlState struct {
	inString             bool
	escaped              bool
	stringBuffer         strings.Builder
	sqlStartLine         int
	sqlPattern           *regexp.Regexp
	relPath              string
	insideQueryCall      bool
	queryBuffer          strings.Builder
	queryLineStart       int
	queryAlreadyCaptured bool
}
