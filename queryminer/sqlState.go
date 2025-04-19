package queryminer

import (
	"regexp"
	"strings"
)

type CaptureType int

const (
	CaptureNone     CaptureType = iota
	CaptureFunction             // * e.g. ->query(...) o ::prepare(...)
	CaptureVariable             // * e.g. $sql = "..." o $query = '...'
)

var (
	// * Detects SQL string fragmentation caused by PHP string concatenation (e.g. " . $var . ").
	phpConcatPattern = regexp.MustCompile(`["']\s*\.\s*|\s*\.\s*["']`)

	phpQueryAssignmentStartPattern = regexp.MustCompile(`(?i)\$\w+\s*=\s*["']`)
	phpVariablesInsideSqlPattern   = regexp.MustCompile(`(?:(?:'?\$)(\w+)(?:'?)?)`)
	queryCallPattern               = regexp.MustCompile(`(?i)(->|::)(query|exec|prepare)\s*\(`)
	sqlPattern                     = regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE)\s+`)
)

type sqlState struct {
	relPath string

	captureType      CaptureType
	insideQuery      bool
	stringStartLine  int
	parenthesisDepth int

	escaped   bool // * True if the previous character was a \ escaping the current one.
	quoteChar byte // * Current quote delimiter (" or ').
	inString  bool

	stringBuffer strings.Builder
}

func newSQLState(relPath string) *sqlState {
	return &sqlState{
		relPath:     relPath,
		captureType: CaptureNone,
	}
}
