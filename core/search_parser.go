package core

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// List of file extensions that I've encountered.
// Some of them aren't eBooks, but they were returned
// in previous search results.
var fileTypes = [...]string{
	"epub",
	"mobi",
	"azw3",
	"html",
	"rtf",
	"pdf",
	"cdr",
	"lit",
	"cbr",
	"doc",
	"htm",
	"jpg",
	"txt",
	"rar", // Compressed extensions should always be last 2 items
	"zip",
}

// sizeUnits are the units a server may separate from the file size with a
// space, keyed in lower case.
var sizeUnits = map[string]bool{
	"b": true, "kb": true, "mb": true, "gb": true, "tb": true,
	"kib": true, "mib": true, "gib": true, "tib": true,
}

// BookDetail contains the details of a single Book found on the IRC server
type BookDetail struct {
	Server string `json:"server"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Format string `json:"format"`
	Size   string `json:"size"`
	Full   string `json:"full"`
}

type ParseError struct {
	Line  string `json:"line"`
	Error error  `json:"error"`
}

func (p *ParseError) MarshalJSON() ([]byte, error) {
	item := struct {
		Line  string `json:"line"`
		Error string `json:"error"`
	}{
		Line:  p.Line,
		Error: p.Error.Error(),
	}
	return json.Marshal(item)
}

func (p ParseError) String() string {
	return fmt.Sprintf("Error: %s. Line: %s.", p.Error, p.Line)
}

// ParseSearchFile converts a single search file into an array of BookDetail
func ParseSearchFile(filePath string) ([]BookDetail, []ParseError, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	books, errs := ParseSearchV2(file)
	return books, errs, nil
}

func ParseSearchV2(reader io.Reader) ([]BookDetail, []ParseError) {
	books := make([]BookDetail, 0)
	parseErrors := make([]ParseError, 0)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "!") {
			dat, err := parseLineV2(line)
			if err != nil {
				parseErrors = append(parseErrors, ParseError{Line: line, Error: err})
			} else {
				books = append(books, dat)
			}
		}
	}

	sort.Slice(books, func(i, j int) bool { return books[i].Server < books[j].Server })

	return books, parseErrors
}

func parseLineV2(line string) (BookDetail, error) {
	getServer := func(line string) (string, error) {
		if len(line) == 0 || line[0] != '!' {
			return "", errors.New("result lines must start with '!'")
		}

		firstSpace := strings.Index(line, " ")
		if firstSpace == -1 {
			return "", errors.New("unable to parse server name")
		}

		return line[1:firstSpace], nil
	}

	// Returns the author along with the index the title starts at.
	//
	// Plenty of servers post files with no author in the name at all, either
	// because the book has none ("!dragnbreaker Inferno! 030 [Black Library]
	// (2002).cbr") or because the whole name is underscore separated
	// ("!DeathCookie Travis_Bagwell_Tarot_03_Inferno.epub.rar"). Those results
	// are perfectly downloadable, so the author is simply left empty and the
	// title takes everything after the server name, rather than the line being
	// discarded as unparseable.
	getAuthor := func(line string) (string, int) {
		firstSpace := strings.Index(line, " ")
		dashChar := strings.Index(line, " - ")
		if dashChar == -1 {
			return "", firstSpace + len(" ")
		}

		author := line[firstSpace+len(" ") : dashChar]
		titleStart := dashChar + len(" - ")

		// Handles case with weird author characters %\w% ("%F77FE9FF1CCD% Michael Haag")
		if strings.Contains(author, "%") {
			if split := strings.SplitAfterN(author, " ", 2); len(split) == 2 {
				author = split[1]
			}
		}

		return author, titleStart
	}

	getTitle := func(line string, titleStart int) (string, string, int) {
		title := ""
		fileFormat := ""
		endIndex := -1
		// Get the Title
		for _, ext := range fileTypes { //Loop through each possible file extension we've got on record
			endTitle := strings.Index(line, "."+ext) // check if it contains our extension
			// Also skips an extension that turns up before the title begins.
			if endTitle < titleStart {
				continue
			}
			fileFormat = ext
			title = line[titleStart:endTitle]
			if ext == "rar" || ext == "zip" { // If the extension is .rar or .zip the actual format is contained in ()
				for _, ext2 := range fileTypes[:len(fileTypes)-2] { // Range over the eBook formats (exclude archives)
					if strings.Contains(strings.ToLower(line[:endTitle]), ext2) {
						fileFormat = ext2
					}
				}
				// "Name.epub.rar": the inner extension is the format, not part
				// of the title.
				title = strings.TrimSuffix(title, "."+fileFormat)
			}
			endIndex = endTitle
		}

		return title, fileFormat, endIndex
	}

	getSize := func(line string) (string, int) {
		const delimiter = " ::INFO:: "
		infoIndex := strings.LastIndex(line, delimiter)
		if infoIndex == -1 {
			return "N/A", len(line)
		}

		// Handle cases when there is additional info after the file size (ex ::HASH:: )
		parts := strings.Split(line[infoIndex+len(delimiter):], " ")
		size := parts[0]

		// Some servers separate the number from its unit ("205.10 KiB") rather
		// than joining them ("8.44MB"), which otherwise loses the unit.
		if len(parts) > 1 && sizeUnits[strings.ToLower(parts[1])] {
			size += " " + parts[1]
		}

		return size, infoIndex
	}

	server, err := getServer(line)
	if err != nil {
		return BookDetail{}, err
	}

	author, titleStart := getAuthor(line)

	title, format, titleIndex := getTitle(line, titleStart)
	if titleIndex == -1 {
		return BookDetail{}, errors.New("unable to parse title")
	}

	size, endIndex := getSize(line)

	return BookDetail{
		Server: server,
		Author: author,
		Title:  title,
		Format: format,
		Size:   size,
		Full:   strings.TrimSpace(line[:endIndex]),
	}, nil
}
