#!/bin/sh

# Print the CHANGELOG.md entry for one version, to use as release notes.
# The file uses a setext heading: the bare version, then a line of dashes.

set -e

version="$1"
if [ -z "$version" ]; then
	echo "usage: $0 <version>" >&2
	exit 1
fi

notes=$(awk -v version="$version" '
	found {
		if ($0 ~ /^---+$/ && previous != "") { previous = ""; exit }
		if (previous != "") print previous
		previous = $0
		next
	}
	$0 ~ /^---+$/ && previous == version {
		found = 1
		previous = ""
		next
	}
	{ previous = $0 }
	END { if (found && previous != "") print previous }
' CHANGELOG.md)

if [ -z "$notes" ]; then
	echo "no CHANGELOG.md entry for version $version" >&2
	exit 1
fi

echo "$notes"
