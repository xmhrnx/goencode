#!/usr/bin/env bash
# Prepares a GitHub Linux runner to package GoEncode for Android; giode.toml's
# [ci.setup] runs it before every Linux build. It installs what the runner
# image lacks (nasm for ffmpeg, bundletool for the .aab) and builds ffmpeg
# for Android (build-ffmpeg.sh).
set -euo pipefail

# The test job doesn't package anything.
if [[ "${GITHUB_JOB:-}" == "test" ]]; then
	echo "ci-setup: nothing to do in the test job"
	exit 0
fi

BUNDLETOOL_VERSION="${BUNDLETOOL_VERSION:-1.18.3}"

sudo apt-get install -y nasm

# giode build aab runs bundletool*.jar from the newest build-tools folder.
SDK="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
BUILD_TOOLS="$(ls -d "$SDK"/build-tools/*/ | sort -V | tail -n 1)"
if ! ls "$BUILD_TOOLS"bundletool*.jar >/dev/null 2>&1; then
	sudo curl -fsSL -o "${BUILD_TOOLS}bundletool.jar" \
		"https://github.com/google/bundletool/releases/download/$BUNDLETOOL_VERSION/bundletool-all-$BUNDLETOOL_VERSION.jar"
	echo "ci-setup: bundletool $BUNDLETOOL_VERSION in $BUILD_TOOLS"
fi

bash "$(dirname "$0")/build-ffmpeg.sh"
