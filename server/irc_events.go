package server

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/evan-buss/openbooks/core"
	"github.com/evan-buss/openbooks/dcc"
)

func (s *server) NewIrcEventHandler() core.EventHandler {
	handler := core.EventHandler{}
	handler[core.SearchResult] = s.searchResultHandler(s.config.DownloadDir)
	handler[core.BookResult] = s.bookResultHandler(s.config.DownloadDir, s.config.DisableBrowserDownloads)
	handler[core.NoResults] = s.noResultsHandler
	handler[core.BadServer] = s.badServerHandler
	handler[core.SearchAccepted] = s.searchAcceptedHandler
	handler[core.MatchesFound] = s.matchesFoundHandler
	handler[core.Ping] = s.pingHandler
	handler[core.ServerList] = s.userListHandler(s.repository)
	handler[core.Version] = s.versionHandler(s.config.UserAgent)
	return handler
}

// searchResultHandler downloads from DCC server, parses data, and sends data to client
func (s *server) searchResultHandler(downloadDir string) core.HandlerFunc {
	return func(text string) {
		extractedPath, err := s.downloadDCC(filepath.Join(downloadDir, "books"), text, "search results")
		if err != nil {
			s.send <- newErrorResponse("Error when downloading search results.")
			return
		}

		bookResults, parseErrors, err := core.ParseSearchFile(extractedPath)
		if err != nil {
			s.log.Println(err)
			s.send <- newErrorResponse("Error when parsing search results.")
			return
		}

		if len(bookResults) == 0 && len(parseErrors) == 0 {
			s.noResultsHandler(text)
			return
		}

		// Output all errors so parser can be improved over time
		if len(parseErrors) > 0 {
			s.log.Printf("%d Search Result Parsing Errors\n", len(parseErrors))
			for _, err := range parseErrors {
				s.log.Println(err)
			}
		}

		query := queryFromResultsFilename(extractedPath)

		s.log.Printf("Sending %d search results for %q.\n", len(bookResults), query)
		s.send <- newSearchResponse(bookResults, parseErrors, query)

		err = os.Remove(extractedPath)
		if err != nil {
			s.log.Printf("Error deleting search results file: %v", err)
		}
	}
}

// bookResultHandler downloads the book file and sends it over the websocket
func (s *server) bookResultHandler(downloadDir string, disableBrowserDownloads bool) core.HandlerFunc {
	return func(text string) {
		extractedPath, err := s.downloadDCC(filepath.Join(downloadDir, "books"), text, "book")
		if err != nil {
			s.send <- newErrorResponse("Error when downloading book.")
			return
		}

		s.log.Printf("Sending book entitled '%s'.\n", filepath.Base(extractedPath))
		s.send <- newDownloadResponse(extractedPath, disableBrowserDownloads)
	}
}

// downloadDCC performs a DCC transfer, logging the address it dials and how
// the transfer ended.
//
// A DCC SEND names an address and a size chosen by whoever is sending, and
// openbooks has to make the outbound connection. When a transfer fails or
// stalls, that address and the byte count are what explain why, and they
// appear nowhere else, so they are logged whether or not raw IRC logging is
// enabled.
func (s *server) downloadDCC(baseDir, dccString, what string) (string, error) {
	if download, err := dcc.ParseString(dccString); err == nil {
		s.log.Printf("Downloading %s: %q from %s:%d, %d bytes.\n",
			what, download.Filename, download.IP, download.Port, download.Size)
	} else {
		s.log.Printf("Unable to parse the DCC SEND for %s, attempting it anyway: %v\n", what, err)
	}

	start := time.Now()

	extractedPath, err := core.DownloadExtractDCCString(baseDir, dccString, nil)
	if err != nil {
		s.log.Printf("Download of %s failed after %s: %v\n", what, time.Since(start).Round(time.Millisecond), err)
		return "", err
	}

	s.log.Printf("Download of %s completed in %s.\n", what, time.Since(start).Round(time.Millisecond))
	return extractedPath, nil
}

// NoResults is called when the server returns that nothing was found for the query
func (s *server) noResultsHandler(_ string) {
	s.send <- newErrorResponse("No results found for the query.")
}

// BadServer is called when the requested download fails because the server is not available
func (s *server) badServerHandler(_ string) {
	s.send <- newErrorResponse("Server is not available. Try another one.")
}

// SearchAccepted is called when the user's query is accepted into the search queue
func (s *server) searchAcceptedHandler(_ string) {
	s.send <- newStatusResponse(NOTIFY, "Search accepted into the queue.")
}

// MatchesFound is called when the server finds matches for the user's query
func (s *server) matchesFoundHandler(num string) {
	s.send <- newStatusResponse(NOTIFY, fmt.Sprintf("Found %s results for your query.", num))
}

func (s *server) pingHandler(serverUrl string) {
	s.irc.Pong(serverUrl)
}

func (s *server) versionHandler(version string) core.HandlerFunc {
	return func(line string) {
		s.log.Printf("Sending CTCP version response: %s", line)
		core.SendVersionInfo(s.irc, line, version)
	}
}

func (s *server) userListHandler(repo *Repository) core.HandlerFunc {
	return func(text string) {
		repo.servers = core.ParseServers(text)
	}
}
