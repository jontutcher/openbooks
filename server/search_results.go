package server

import (
	"path/filepath"
	"regexp"
	"strings"
)

// The search bot names the results file after the query it answered, as in
// "SearchBot_results_for__dune.txt.zip". Which bot replied is a configuration
// detail and the separator has changed before — the mock fixture uses a space
// where IRC Highway now sends an underscore — so match the "results for" part
// rather than any one bot's name, and treat both separators alike.
var resultsForPattern = regexp.MustCompile(`(?i)^.*results[_\s]+for[_\s]*`)

// queryFromResultsFilename recovers the query a results file answers from its
// name. It returns "" when the name does not look like a results file, which
// leaves the client to fall back to its own ordering rather than attributing
// the results to the wrong search.
func queryFromResultsFilename(path string) string {
	name := filepath.Base(path)

	// Both ".txt.zip" and the ".txt" left after extraction.
	for ext := filepath.Ext(name); ext != ""; ext = filepath.Ext(name) {
		name = strings.TrimSuffix(name, ext)
	}

	if !resultsForPattern.MatchString(name) {
		return ""
	}

	return normalizeQuery(resultsForPattern.ReplaceAllString(name, ""))
}

// normalizeQuery puts a query in the form the client compares on: the bot
// replaces the spaces in a query with underscores, so neither separator can be
// taken literally on either side of the comparison.
func normalizeQuery(query string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return r == '_' || r == ' ' || r == '\t'
	}), " ")
}
