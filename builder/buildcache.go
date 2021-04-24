package builder

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tinygo-org/tinygo/goenv"
)

// Return the newest timestamp of all the file paths passed in. Used to check
// for stale caches.
func cacheTimestamp(paths []string) (time.Time, error) {
	var timestamp time.Time
	for _, path := range paths {
		st, err := os.Stat(path)
		if err != nil {
			return time.Time{}, err
		}
		if timestamp.IsZero() {
			timestamp = st.ModTime()
		} else if timestamp.Before(st.ModTime()) {
			timestamp = st.ModTime()
		}
	}
	return timestamp, nil
}

// Try to load a given file from the cache. Return "", nil if no cached file can
// be found (or the file is stale), return the absolute path if there is a cache
// and return an error on I/O errors.
func cacheLoad(name string, sourceFiles []string) (string, error) {
	cachepath := filepath.Join(goenv.Get("GOCACHE"), name)
	cacheStat, err := os.Stat(cachepath)
	if os.IsNotExist(err) {
		return "", nil // does not exist
	} else if err != nil {
		return "", err // cannot stat cache file
	}

	sourceTimestamp, err := cacheTimestamp(sourceFiles)
	if err != nil {
		return "", err // cannot stat source files
	}

	if cacheStat.ModTime().After(sourceTimestamp) {
		return cachepath, nil
	} else {
		os.Remove(cachepath)
		// stale cache
		return "", nil
	}
}

// Store the file located at tmppath in the cache with the given name. The
// tmppath may or may not be gone afterwards.
func cacheStore(tmppath, name string, sourceFiles []string) (string, error) {
	// get the last modified time
	if len(sourceFiles) == 0 {
		panic("cache: no source files")
	}

	// TODO: check the config key

	dir := goenv.Get("GOCACHE")
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return "", err
	}
	cachepath := filepath.Join(dir, name)
	err = copyFile(tmppath, cachepath)
	if err != nil {
		return "", err
	}
	return cachepath, nil
}

// copyFile copies the given file from src to dst. It can copy over
// a possibly already existing file at the destination.
func copyFile(src, dst string) error {
	inf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inf.Close()
	outpath := src + ".tmp"
	outf, err := os.Create(outpath)
	if err != nil {
		return err
	}

	_, err = io.Copy(outf, inf)
	if err != nil {
		os.Remove(outpath)
		return err
	}

	err = outf.Close()
	if err != nil {
		return err
	}

	// Rename may fail if another process is trying to write to
	// the same file.  However, in this case, the failure is
	// acceptable because the result of the other process can be
	// used.
	os.Rename(outpath, dst)
	return nil
}
