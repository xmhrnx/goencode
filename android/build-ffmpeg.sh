#!/usr/bin/env bash
# Builds ffmpeg and ffprobe for Android with the NDK and packs them into
# android/ffmpeg.aar as jni/<abi>/libffmpeg.so and libffprobe.so, which
# giode merges into the APK (giode.toml [android] libraries). Android only
# lets apps run programs from their native library folder, and only files
# named lib*.so get extracted there; see docs/ffmpeg.md.
#
# Needs: Linux or macOS, the Android NDK (ANDROID_NDK_HOME, or the
# ANDROID_NDK_LATEST_HOME GitHub's runners set), make, curl, git, zip and
# nasm. Takes 15-30 minutes. Run it once before `giode build android`.
#
# The build is GPL (it includes x264), so the APK is too.
set -euo pipefail

# CI runs giode.toml's [ci.setup] in every Linux job; only the ones that
# package Android need ffmpeg.
if [[ "${GITHUB_JOB:-}" == "test" ]]; then
	echo "build-ffmpeg: nothing to do in the test job"
	exit 0
fi

FFMPEG_VERSION="${FFMPEG_VERSION:-7.1.1}"
API="${ANDROID_API:-24}" # giode.toml android.min_sdk
ABIS="${ABIS:-arm64-v8a armeabi-v7a x86_64}"

HERE="$(cd "$(dirname "$0")" && pwd)"
OUT="$HERE/ffmpeg.aar"
WORK="${WORK:-$HERE/.ffmpeg-build}"

if [[ -f "$OUT" && -z "${FORCE:-}" ]]; then
	echo "build-ffmpeg: $OUT exists (FORCE=1 rebuilds it)"
	exit 0
fi

NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK_LATEST_HOME:-${ANDROID_NDK_ROOT:-}}}"
if [[ -z "$NDK" || ! -d "$NDK" ]]; then
	echo "build-ffmpeg: set ANDROID_NDK_HOME to the Android NDK" >&2
	exit 1
fi
case "$(uname -s)" in
Linux) HOST=linux-x86_64 ;;
Darwin) HOST=darwin-x86_64 ;;
*) echo "build-ffmpeg: run this on Linux or macOS" >&2; exit 1 ;;
esac
TOOLS="$NDK/toolchains/llvm/prebuilt/$HOST"
SYSROOT="$TOOLS/sysroot"
JOBS="$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)"

mkdir -p "$WORK"
cd "$WORK"
if [[ ! -d "ffmpeg-$FFMPEG_VERSION" ]]; then
	curl -fsSL "https://ffmpeg.org/releases/ffmpeg-$FFMPEG_VERSION.tar.xz" | tar xJ
fi
if [[ ! -d x264 ]]; then
	git clone --depth 1 --branch stable https://code.videolan.org/videolan/x264.git
fi

STAGE="$WORK/aar"
rm -rf "$STAGE"
mkdir -p "$STAGE"

for ABI in $ABIS; do
	case "$ABI" in
	arm64-v8a) TRIPLE=aarch64-linux-android; ARCH=aarch64; CPU=armv8-a; X264_HOST=aarch64-linux-android ;;
	armeabi-v7a) TRIPLE=armv7a-linux-androideabi; ARCH=arm; CPU=armv7-a; X264_HOST=arm-linux-androideabi ;;
	x86_64) TRIPLE=x86_64-linux-android; ARCH=x86_64; CPU=x86-64; X264_HOST=x86_64-linux-android ;;
	*) echo "build-ffmpeg: unknown ABI $ABI" >&2; exit 1 ;;
	esac
	echo "==> $ABI"
	CC="$TOOLS/bin/$TRIPLE$API-clang"
	CXX="$TOOLS/bin/$TRIPLE$API-clang++"
	PREFIX="$WORK/prefix-$ABI"

	# x264, static, for libx264 in ffmpeg.
	rm -rf "build-x264-$ABI" && cp -r x264 "build-x264-$ABI"
	(
		cd "build-x264-$ABI"
		CC="$CC" AR="$TOOLS/bin/llvm-ar" RANLIB="$TOOLS/bin/llvm-ranlib" STRIP="$TOOLS/bin/llvm-strip" \
			./configure --host="$X264_HOST" --cross-prefix="$TOOLS/bin/llvm-" --sysroot="$SYSROOT" \
			--prefix="$PREFIX" --enable-static --enable-pic --disable-cli \
			$([[ "$ABI" == armeabi-v7a ]] && echo --disable-asm)
		make -j"$JOBS" && make install
	)

	# ffmpeg and ffprobe as position-independent executables, statically
	# linked against everything but Android's system libraries.
	rm -rf "build-ffmpeg-$ABI" && mkdir "build-ffmpeg-$ABI"
	(
		cd "build-ffmpeg-$ABI"
		PKG_CONFIG_PATH="$PREFIX/lib/pkgconfig" PKG_CONFIG_LIBDIR="$PREFIX/lib/pkgconfig" \
			"../ffmpeg-$FFMPEG_VERSION/configure" \
			--target-os=android --enable-cross-compile --arch="$ARCH" --cpu="$CPU" \
			--cc="$CC" --cxx="$CXX" --sysroot="$SYSROOT" \
			--nm="$TOOLS/bin/llvm-nm" --ar="$TOOLS/bin/llvm-ar" --ranlib="$TOOLS/bin/llvm-ranlib" \
			--strip="$TOOLS/bin/llvm-strip" --pkg-config=pkg-config --pkg-config-flags=--static \
			--extra-cflags="-I$PREFIX/include -fPIE" --extra-ldflags="-L$PREFIX/lib -fPIE -pie" \
			--enable-pic --enable-static --disable-shared --disable-doc --disable-ffplay \
			--disable-debug --enable-gpl --enable-version3 --enable-libx264 \
			--enable-jni --enable-mediacodec \
			$([[ "$ABI" == x86_64 ]] && echo --disable-x86asm)
		make -j"$JOBS"
		mkdir -p "$STAGE/jni/$ABI"
		"$TOOLS/bin/llvm-strip" -o "$STAGE/jni/$ABI/libffmpeg.so" ffmpeg
		"$TOOLS/bin/llvm-strip" -o "$STAGE/jni/$ABI/libffprobe.so" ffprobe
	)
done

cat >"$STAGE/AndroidManifest.xml" <<'XML'
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.goencode.ffmpeg" />
XML
rm -f "$OUT"
(cd "$STAGE" && zip -qr "$OUT" .)
echo "build-ffmpeg: wrote $OUT"
