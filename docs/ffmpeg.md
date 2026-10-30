# ffmpeg on each platform

GoEncode starts `ffmpeg` and `ffprobe` as separate programs, so it doesn't need cgo and can use
any ffmpeg build. What a build can do depends on how it was configured. For example, x265,
SVT‑AV1 and NVENC are only offered when the ffmpeg build includes them. A "full" build gives
the most choice.

## Windows, macOS, Linux

ffmpeg is looked up in this order:

1. The location set in **Preferences** (the program or its folder).
2. The folder that holds the GoEncode executable.
3. Common package-manager folders.
4. `PATH`.

To ship ffmpeg with GoEncode, put `ffmpeg` and `ffprobe` (`.exe` on Windows) next to the
executable inside the package. Check that the ffmpeg build's license allows redistribution
alongside your app.

## Android

Android only lets an app run programs from its *native library folder*. That folder holds the
`.so` files packaged in the APK and is the only place an app can execute files from.

`android/build-ffmpeg.sh` handles this. Run it on Linux or macOS with `ANDROID_NDK_HOME` set,
before `giode build android`. The release workflow runs it automatically through `[ci.setup]`.
It builds ffmpeg 7.1.1 with x264, MediaCodec and ffmpeg's built-in encoders for `arm64-v8a`,
`armeabi-v7a` and `x86_64`. The x264 build is GPL, so the APK is too. Set `ABIS` or
`FFMPEG_VERSION` to change what it builds.

The script follows these steps, which you can also do by hand with your own ffmpeg build:

1. Build ffmpeg and ffprobe as executables for each ABI with the NDK, and name them
   `libffmpeg.so` and `libffprobe.so`. The `lib…so` names are required, because the installer
   only extracts files named like that.
2. Put them in an `.aar`:

   ```
   ffmpeg.aar
   ├── AndroidManifest.xml      <manifest package="com.goencode.ffmpeg"/>
   └── jni/
       └── arm64-v8a/
           ├── libffmpeg.so
           └── libffprobe.so
   ```

3. List the `.aar` in `giode.toml`. giode merges its `.so` files into the APK, and because the
   manifest doesn't set `extractNativeLibs="false"`, Android unpacks them to disk on install:

   ```toml
   [android]
   libraries = ["android/ffmpeg.aar"]
   ```

At run time, `internal/ffmpeg/binary_android.go` finds the native library folder from the app's
own memory map, and `Locate` runs ffmpeg from there.

Files picked on Android come in as streams. GoEncode copies them into its own storage, encodes
them there, and offers results through the system share sheet.

## iOS

iOS doesn't allow an app to start other programs, so GoEncode can't run ffmpeg there. On iOS it
shows a notice explaining this, and `giode.toml` leaves iOS out of the release platforms. iOS
support would need ffmpeg linked as a library through cgo (the libav* APIs) instead of the CLI.
That is a separate engine behind the same `settings.Args` recipe.
