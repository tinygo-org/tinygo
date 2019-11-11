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
//
// TODO: the configKey is currently ignored. It is supposed to be used as extra
// data for the cache key, like the compiler version and arguments.
func cacheLoad(name, configKey string, sourceFiles []string) (string, error) {
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
//
// Note: the configKey is ignored, see cacheLoad.
func cacheStore(tmppath, name, configKey string, sourceFiles []string) (string, error) {
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
	err = moveFile(tmppath, cachepath)
	if err != nil {
		return "", err
	}
	return cachepath, nil
}

// moveFile renames the file from src to dst. If renaming doesn't work (for
// example, the rename crosses a filesystem boundary), the file is copied and
// the old file is removed.
func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		// Success!
		return nil
	}
	// Failed to move, probably a different filesystem.
	// Do a copy + remove.
	inf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inf.Close()
	outpath := dst + ".tmp"
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

	return os.Rename(dst+".tmp", dst)
}
