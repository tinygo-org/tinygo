package main

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

// Get the cache directory, usually ~/.cache/tinygo
func cacheDir() string {
	home := getHomeDir()
	dir := filepath.Join(home, ".cache", "tinygo")
	return dir
}

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
	dir := cacheDir()
	cachepath := filepath.Join(dir, name)
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

	dir := cacheDir()
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return "", err
	}
	cachepath := filepath.Join(dir, name)
	err = os.Rename(tmppath, cachepath)
	if err != nil {
		inf, err := os.Open(tmppath)
		if err != nil {
			return "", err
		}
		defer inf.Close()
		outf, err := os.Create(cachepath + ".tmp")
		if err != nil {
			return "", err
		}

		_, err = io.Copy(outf, inf)
		if err != nil {
			return "", err
		}

		err = os.Rename(cachepath+".tmp", cachepath)
		if err != nil {
			return "", err
		}

		return cachepath, outf.Close()
	}
	return cachepath, nil
}
