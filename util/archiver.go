package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

var (
	ErrNotFullyCopied = errors.New("didn't copy entire file from the archive")
)

// tempSuffix is appended to in-progress downloads. It has to be removed before
// archive detection so that the real file extension is visible.
const tempSuffix = ".temp"

// originalName strips the ".temp" suffix that in-progress downloads carry.
func originalName(path string) string {
	return strings.TrimSuffix(path, tempSuffix)
}

// ExtractArchive extracts the single file contained in the archive at
// archivePath and returns its path. If the archive holds anything other than
// exactly one file, the archive itself is delivered instead and its path is
// returned unchanged.
func ExtractArchive(archivePath string) (string, error) {
	ctx := context.Background()

	archive, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer archive.Close()

	// The path has ".temp" appended to it, so detection is given the original
	// name; passing the open file also lets the format be matched on content.
	format, stream, err := archives.Identify(ctx, originalName(archivePath), archive)
	if err != nil {
		return "", err
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return "", fmt.Errorf("format of %s (%s) does not support extraction", archivePath, format.Extension())
	}

	var newPath string
	err = extractor.Extract(ctx, stream, func(ctx context.Context, f archives.FileInfo) error {
		// Directory entries carry no content of their own.
		if f.IsDir() {
			return nil
		}

		// Extract only one file per archive. Otherwise, stop extracting,
		// remove the extracted item, and deliver the archive itself.
		if newPath != "" {
			if err := os.Remove(newPath); err != nil {
				return err
			}
			newPath = ""
			return fs.SkipAll
		}

		// Never trust a name from inside the archive; only its base name is
		// used, so that an entry such as "../../evil" cannot escape the
		// download directory.
		name := filepath.Base(filepath.FromSlash(f.NameInArchive))
		if name == "." || name == string(filepath.Separator) {
			name = f.Name()
		}
		newPath = filepath.Join(filepath.Dir(archivePath), name+tempSuffix)

		entry, err := f.Open()
		if err != nil {
			return err
		}
		defer entry.Close()

		out, err := os.Create(newPath)
		if err != nil {
			return err
		}
		defer out.Close()

		copied, err := io.Copy(out, entry)
		if err != nil {
			return err
		}
		if copied != f.Size() {
			return ErrNotFullyCopied
		}

		return out.Close()
	})

	if err != nil {
		return "", err
	}

	// If we extracted exactly one file, send that file and remove the archive.
	// Otherwise, send the archive itself.
	if newPath != "" {
		if err := os.Remove(archivePath); err != nil {
			log.Println("remove error", err)
		}
		return newPath, nil
	}

	return archivePath, nil
}

// IsArchive returns true if the file at the given path is an archive that can
// be extracted. Returns false otherwise.
func IsArchive(path string) bool {
	// Matching is done on the name alone so that the file does not have to be
	// read; ExtractArchive confirms the format against the contents later.
	format, _, err := archives.Identify(context.Background(), originalName(path), nil)
	if err != nil {
		return false
	}

	// Formats that only compress a single stream (".gz", ".bz2", ...) have
	// nothing to walk, so they are delivered as they arrived.
	_, ok := format.(archives.Extractor)
	return ok
}
