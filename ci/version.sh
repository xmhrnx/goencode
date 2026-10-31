#!/usr/bin/env bash
# Makes the release tag a version giode accepts; giode.toml's [ci.setup]
# runs it before every build. giode takes major.minor.patch and fails on
# anything else, so a tag like v1.0.0.0 keeps its first three parts and
# v1.2 becomes 1.2.0. The build number (Android's versionCode, the
# fourth part on Windows) comes from the run counter, $GIODE_BUILD.
set -euo pipefail

v="${GIODE_VERSION:-}"
v="${v#[vV]}"
if [[ -z "$v" ]]; then
	exit 0
fi
if [[ ! "$v" =~ ^[0-9]+(\.[0-9]+)*$ ]]; then
	echo "ci/version.sh: the tag $GIODE_VERSION isn't a version like v1.2.3" >&2
	exit 1
fi

IFS=. read -r -a parts <<<"$v"
while ((${#parts[@]} < 3)); do
	parts+=(0)
done
version="$((10#${parts[0]})).$((10#${parts[1]})).$((10#${parts[2]}))"

echo "ci/version.sh: building version $version (tag $GIODE_VERSION)"
if [[ -n "${GITHUB_ENV:-}" ]]; then
	echo "GIODE_VERSION=$version" >>"$GITHUB_ENV"
fi
