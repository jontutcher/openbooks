package server

import "testing"

func TestQueryFromResultsFilename(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		// The names IRC Highway actually sent, taken from a container log.
		{"underscore separator", "SearchBot_results_for__dune.txt.zip", "dune"},
		{"multi word", "SearchBot_results_for__foundation_asimov.txt.zip", "foundation asimov"},
		{"after extraction", "/books/books/SearchBot_results_for__florida.txt", "florida"},
		// The name the mock in cmd/mock serves, which uses spaces.
		{"space separator", "SearchBot_results_for_ the great gatsby.txt.zip", "the great gatsby"},
		{"another bot", "Search_results_for_dune.txt", "dune"},
		{"mixed case", "SEARCHBOT_RESULTS_FOR__Dune.TXT", "dune"},
		// Not a results file: the client must fall back rather than be told
		// the results belong to a query they do not.
		{"book file", "/books/books/Frank Herbert - Dune.epub", ""},
		{"empty", "", ""},
		{"prefix only", "SearchBot_results_for_.txt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queryFromResultsFilename(tt.path); got != tt.want {
				t.Errorf("queryFromResultsFilename(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"Dune", "dune"},
		{"foundation_asimov", "foundation asimov"},
		{"  the   great  gatsby ", "the great gatsby"},
		{"Neuromancer", "neuromancer"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := normalizeQuery(tt.query); got != tt.want {
			t.Errorf("normalizeQuery(%q) = %q, want %q", tt.query, got, tt.want)
		}
	}
}
