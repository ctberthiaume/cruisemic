package storage

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

// Storer is the interface that wraps methods to store data feeds as text.
type Storer interface {
	WriteString(feed string, s string) error
	Close() error
	Flush() error
	FeedPath(feed string) string
}

type MemStorage struct {
	Feeds   map[string][]string
	Flushed bool
	Closed  bool
}

// NewMemStorage creates a new MemStorage struct.
func NewMemStorage() (*MemStorage, error) {
	f := make(map[string][]string)
	return &MemStorage{Feeds: f}, nil
}

// WriteString stores a string for a feed.
func (store *MemStorage) WriteString(feed string, s string) (err error) {
	store.Feeds[feed] = append(store.Feeds[feed], s)
	return
}

// Flush simulates commiting data to persistent storage. It sets Flushed to
// true.
func (store *MemStorage) Flush() (err error) {
	store.Flushed = true
	return
}

// Close simulates closing open persistent storage resources. It sets Closed to
// true.
func (store *MemStorage) Close() (err error) {
	store.Closed = true
	return
}

// FeedPath creates a feed file path. For MemStorage this is just the feed name.
func (store *MemStorage) FeedPath(feed string) string {
	return feed
}

// DiskStorage implements methods to save text data feeds to disk.
type DiskStorage struct {
	dir        string
	filePrefix string
	fileExt    string
	files      map[string]*os.File
	out        map[string]*bufio.Writer
	buffSize   int
}

// NewDiskStorage creates a new DiskStorage struct. Data will be written to
// files in dir, with names <filePrefix><feed><ext>. Extension <ext> should
// contain a leading dot. feeds should be used to declare any feed files that
// will be written too, and to associate feed names with any header text
// to be written. Header text will only be written if the file is empty.
func NewDiskStorage(dir string, filePrefix string, fileExt string, feedHeaders map[string]string, buffSize int) (*DiskStorage, error) {
	if buffSize <= 0 {
		buffSize = 1 << 16 // 65536
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	store := &DiskStorage{
		dir:      dir,
		files:    map[string]*os.File{},
		out:      map[string]*bufio.Writer{},
		buffSize: buffSize,
	}
	store.filePrefix = filePrefix
	store.fileExt = fileExt

	// Open feed files and write header if necessary
	for feed, header := range feedHeaders {
		if err := store.writeHeader(feed, header); err != nil {
			return nil, err
		}
	}

	return store, nil
}

// WriteString writes a string to feed output file.
func (store *DiskStorage) WriteString(feed string, s string) error {
	out, ok := store.out[feed]
	if !ok {
		if err := store.setOutput(feed); err != nil {
			return err
		}
		out = store.out[feed]
	}
	_, err := out.WriteString(s)
	return err
}

// Flush flushes all open file resources. This function will always try to
// flush all resources, and if errors occur the last error will be returned.
func (store *DiskStorage) Flush() (err error) {
	for _, v := range store.out {
		if e := v.Flush(); e != nil {
			err = e
		}
	}

	return err
}

// Close flushes and closes all open file resources. This function will always
// try to flush and close all resources, and if errors occur the last error will
// be returned.
func (store *DiskStorage) Close() (err error) {
	err = store.Flush()
	for _, v := range store.files {
		if e := v.Close(); e != nil {
			err = e
		}
	}

	return err
}

// FeedPath creates a feed file path.
func (store *DiskStorage) FeedPath(feed string) string {
	return filepath.Join(store.dir, store.filePrefix+feed+store.fileExt)
}

// hasData checks if the output feed already contains data.
func (store *DiskStorage) hasData(feed string) (bool, error) {
	file, err := os.Open(store.FeedPath(feed))
	if err != nil {
		return false, err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return false, err
	}
	return fi.Size() > 0, nil
}

// setOutput opens an output file for a data feed.
func (store *DiskStorage) setOutput(feed string) error {
	path := store.FeedPath(feed)
	of, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	store.files[feed] = of
	store.out[feed] = bufio.NewWriterSize(of, store.buffSize)
	return nil
}

// writeHeaders writes feed headers. If a feed file already exists and has data,
// no headers will be written.
func (store *DiskStorage) writeHeader(feed string, header string) error {
	_, ok := store.out[feed]
	if !ok {
		if err := store.setOutput(feed); err != nil {
			return err
		}
	}
	hasData, err := store.hasData(feed)
	if err != nil {
		return err
	}
	if !hasData {
		if len(header) > 0 && header[len(header)-1] != "\n"[0] {
			header += "\n"
		}
		if err := store.WriteString(feed, header); err != nil {
			return err
		}
	}
	return nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create a temporary file in the destination directory
	dstDir := filepath.Dir(dst)
	dstBase := filepath.Base(dst)
	tempFile, err := os.CreateTemp(dstDir, dstBase+".tmp-*")
	if err != nil {
		return err
	}
	// Clean up the temp file if something goes wrong
	defer os.Remove(tempFile.Name())

	// Stream the content from source to destination
	_, err = io.Copy(tempFile, sourceFile)
	if err != nil {
		tempFile.Close()
		return err
	}

	// Ensure the file is written to disk and close it
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	// Atomically rename the temporary file to the final destination
	return os.Rename(tempFile.Name(), dst)
}
