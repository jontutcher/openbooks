package util

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// writeZip builds a zip archive at path containing the given entries. An entry
// name ending in "/" is written as a directory entry.
func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestIsArchive(t *testing.T) {
	tests := map[string]bool{
		"book.zip":          true,
		"book.zip.temp":     true,
		"book.rar.temp":     true,
		"book.tar.gz.temp":  true,
		"book.7z.temp":      true,
		"book.epub.temp":    false,
		"book.pdf":          false,
		"book.mobi.temp":    false,
		"book.epub.gz.temp": false, // compression only, nothing to walk
		"no_extension.temp": false,
	}

	for path, want := range tests {
		if got := IsArchive(path); got != want {
			t.Errorf("IsArchive(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestExtractArchiveSingleFile(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "book.zip.temp")
	writeZip(t, archivePath, map[string]string{"book.epub": "contents"})

	got, err := ExtractArchive(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(dir, "book.epub.temp")
	if got != want {
		t.Fatalf("ExtractArchive returned %q, want %q", got, want)
	}

	contents, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "contents" {
		t.Errorf("extracted contents = %q, want %q", contents, "contents")
	}

	// The archive itself is replaced by the file it held.
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Errorf("archive %q still exists after extraction", archivePath)
	}
}

func TestExtractArchiveSingleFileInDirectory(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "book.zip.temp")
	writeZip(t, archivePath, map[string]string{"nested/book.epub": "contents"})

	got, err := ExtractArchive(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	// Only the base name is used, so the nested path cannot escape the
	// download directory.
	want := filepath.Join(dir, "book.epub.temp")
	if got != want {
		t.Fatalf("ExtractArchive returned %q, want %q", got, want)
	}
}

func TestExtractArchiveMultipleFilesDeliversArchive(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "books.zip.temp")
	writeZip(t, archivePath, map[string]string{
		"first.epub":  "one",
		"second.epub": "two",
	})

	got, err := ExtractArchive(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	if got != archivePath {
		t.Fatalf("ExtractArchive returned %q, want the archive %q", got, archivePath)
	}

	// Nothing may be left behind from the aborted extraction.
	remaining, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 || remaining[0].Name() != "books.zip.temp" {
		names := make([]string, 0, len(remaining))
		for _, entry := range remaining {
			names = append(names, entry.Name())
		}
		t.Errorf("directory contains %v, want only the archive", names)
	}
}

func TestExtractArchiveTraversalEntryStaysInDirectory(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "book.zip.temp")
	writeZip(t, archivePath, map[string]string{"../../escaped.epub": "contents"})

	got, err := ExtractArchive(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(dir, "escaped.epub.temp")
	if got != want {
		t.Fatalf("ExtractArchive returned %q, want %q", got, want)
	}
}
