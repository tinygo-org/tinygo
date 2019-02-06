package main

// tinyGoVersion of this package.
// Update this value before release of new version of software.
const tinyGoVersion = "0.1.0"

// version returns the current TinyGo version
func version() string {
	return tinyGoVersion
}

// displayversion returns the current TinyGo full version description for display purposes.
func displayversion() string {
	return "TinyGo compiler v" + tinyGoVersion
}
