# Releasing GoEncode

Publishing a GitHub release with a tag like `v1.0.0` runs
`.github/workflows/release.yml`. The tag sets the version and the workflow's run
number sets the build number (Android's version code). The workflow tests,
builds every package and attaches them to the release. A "Run workflow" button
in the Actions tab builds the same packages without a release, to try them out.

The workflow is written by the Giode CLI from `giode.toml`: after changing `[ci]`
or a store section, run `giode ci -f` instead of editing it, then redo the two
hand edits noted in `[ci]` (pin `giode@v0.1.35-alpha.1`, `go-version-file: go.mod`).

| Where | What the release gets | Needs |
| --- | --- | --- |
| GitHub, Linux | `.deb` and `.rpm`, x64 | nothing |
| GitHub, Windows | `GoEncode-<version>-amd64-setup.exe` | nothing; a code signing certificate stops SmartScreen warnings |
| GitHub, macOS | a universal `.dmg` | a Developer ID certificate and notarization, or macOS refuses to open it |
| GitHub, Android | an `.apk` to install directly | the upload keystore (unsigned builds use a debug key) |
| Microsoft Store | `GoEncode-<version>-amd64.msix` to upload in Partner Center | a Partner Center account and the reserved name |
| Google Play | an `.aab` to upload in the Play Console | a Play Console account and the upload keystore |

**Not offered:**

- **Mac App Store.** Every store app runs in the App Sandbox, which only lets an
  app start programs inside its own bundle, signed with the app's certificate
  and the sandbox's `inherit` entitlement. GoEncode's whole job is running
  ffmpeg: on a Mac that's the user's ffmpeg (Homebrew, MacPorts or a path set in
  Preferences), which the sandbox blocks: it can't even read `/opt/homebrew`, and
  a path typed in Preferences grants no access. A store version would need
  ffmpeg bundled and signed inside the app, which giode's `mas` build can't do
  yet, and the usual ffmpeg builds (with x264/x265) are GPL, which can't be
  distributed under the App Store's terms. App Review would also reject an app
  that does nothing until the user installs another program (guideline 2.4.5).
  Encodes are also written next to their sources, which the sandbox only allows
  for files the user picked, not their folders. Mac users get the notarized
  `.dmg`, which is signed with the hardened runtime but not sandboxed, so it
  runs any ffmpeg.
- **Apple App Store (iOS).** iOS doesn't let apps start other programs at all.
  `app.platforms` leaves iOS out; see [docs/ffmpeg.md](docs/ffmpeg.md).
- **Snap Store and Flathub.** Not set up; Linux users download the `.deb` or `.rpm`.

## Repository secrets

Set these in GitHub under Settings → Secrets and variables → Actions. Each
`*_BASE64` secret is the file encoded with `base64 -i file | pbcopy` on macOS,
or `base64 -w0 file` on Linux. Without a secret, that step is skipped (the
package is unsigned) and the rest of the release still builds.

| Secret | What it is |
| --- | --- |
| `ANDROID_KEYSTORE_BASE64`, `GIODE_KEYSTORE_PASS` | The Google Play upload keystore (`release.keystore`) and its password: signs the `.apk` and `.aab` |
| `MACOS_CERT_BASE64`, `MACOS_CERT_PASS` | The "Developer ID Application" certificate as a `.p12`, and its password: signs the `.dmg` |
| `GIODE_NOTARY_ID`, `GIODE_NOTARY_TEAM_ID`, `GIODE_NOTARY_PASS` | Your Apple ID, team ID and an app-specific password: notarizes the `.dmg` |
| `WINDOWS_CERT_BASE64`, `GIODE_WINDOWS_CERT_PASS` | An Authenticode `.pfx` and its password: signs the Windows installer (optional; the Store signs the MSIX itself) |

No Mac App Store or iOS secrets are needed.

## Microsoft Store

1. Open a Partner Center developer account (the one EasySend and AnyFile use).
2. Reserve the name "GoEncode" (Apps and games → New product → MSIX or PWA app).
   If it's taken, reserve another name and set `app.name` for the MSIX build to
   it, as AnyFile's workflow does.
3. Open Product management → Product identity and check that `[windows.msix]`
   in `giode.toml` matches it; fix the values if they differ:
   - **Package/Identity/Name** → `identity_name` (`MehranZahir.GoEncode`).
   - **Package/Identity/Publisher** → `publisher` (`CN=B40E8F57-A7E0-40BB-8E74-5D051D0B578B`).
   - **Package/Properties/PublisherDisplayName** → `publisher_display_name` (`Mehran Zahiroddini`).
4. Publish a release. Upload `GoEncode-<version>-amd64.msix` in a new
   submission's Packages page. The Store signs it.
5. **Properties:** category Multimedia design (or Photo & video); privacy policy
   URL `https://xmhrnx.github.io/goencode/#privacy`; website
   `https://xmhrnx.github.io/goencode/`.
6. **Age ratings:** answer the IARC questionnaire with no to everything (no
   user-to-user communication, no shared location, no purchases, no
   user-generated content shared with others). Expect 3+ / Everyone.
7. **Restricted capability:** the submission asks why the package uses
   `runFullTrust`. Answer: "GoEncode is a desktop app that starts the ffmpeg and
   ffprobe programs to read and encode the user's media files, and writes the
   results next to them. It needs full trust to run those programs."
8. **Notes for certification** (ffmpeg isn't in the package, so testers must
   install it):

   > GoEncode is a front end for ffmpeg, which it needs to work. Before testing,
   > install ffmpeg with `winget install Gyan.FFmpeg` in a terminal, then start
   > GoEncode. Add any video file with "Add files" or by dropping it on the
   > window, pick a preset and press Start. If ffmpeg isn't installed, GoEncode
   > shows how to install it and a button to point it at ffmpeg.exe.

**Store notes.** The MSIX is a full-trust desktop app, so it finds ffmpeg as the
setup.exe does: Preferences, common folders (`C:\ffmpeg\bin`, Chocolatey,
Scoop) and `PATH` (winget's links folder is on it). One difference: putting
`ffmpeg.exe` next to `GoEncode.exe` doesn't work, because the Store installs
into the read-only `WindowsApps` folder. As with any app started from the Start
menu, a `PATH` change made by installing ffmpeg is seen after signing out and in
(or point GoEncode at ffmpeg in Preferences). The Store's policies ask that a
dependency on software not included be stated in the description; the listing below does. Bundling an
ffmpeg build in the MSIX would remove the dependency, but giode has no setting
for extra program files yet, and it would make the package GPL (or LGPL with a
reduced build); decide that before trying.

## Google Play

1. Create the upload keystore once, and keep it and its password safe:

   ```
   GIODE_KEYSTORE_PASS='a long password' giode keystore
   ```

   It writes `release.keystore`, which `.gitignore` keeps out of git.
   `giode.toml` sets `android.keystore = "release.keystore"`: the CI writes the
   file there from the secret, and `giode build -release` fails without that
   setting, so keep it even though the file isn't committed. Add it as `ANDROID_KEYSTORE_BASE64` and the
   password as `GIODE_KEYSTORE_PASS`.
2. In the Play Console, create the app "GoEncode" (app, free). The package name
   is `com.goencode.app` (`app.id`), fixed once the first bundle is uploaded.
3. Publish a release and upload the `.aab` to Internal testing first. Accept
   **Play App Signing**: Google keeps the app signing key and `release.keystore`
   is your upload key (if you lose it, Play can reset it; without Play App
   Signing you couldn't update the app).
4. **Check ffmpeg in a Play install.** Install the internal test from the Play
   Store on a phone, add a video and encode it. `giode.toml` sets
   `extractNativeLibs="true"` so Android unpacks ffmpeg from the bundle's split
   APKs; if GoEncode says ffmpeg wasn't found, this is what broke.
5. **App content:**
   - **Privacy policy:** `https://xmhrnx.github.io/goencode/#privacy`.
   - **Data safety:** "No, the app doesn't collect or share user data." GoEncode
     has no internet access; picked files are copied into its own storage and
     processed on the device, which Google doesn't count as collection.
   - **Foreground service:** the manifest has `FOREGROUND_SERVICE_DATA_SYNC`
     (giode's keep-alive service). Declare "Data sync": "While the user encodes
     a batch of their media files, a foreground service with a notification
     keeps ffmpeg running when GoEncode is in the background. It starts when the
     user presses Start and stops when the batch ends." Play asks for a short
     video of this: start a batch, leave the app, show the notification.
   - **Content rating:** IARC questionnaire, category Utility/Productivity, no
     to everything. Expect Everyone / PEGI 3.
   - **Target audience:** 13 and over (not designed for children).
   - **Ads:** no ads. **Government app, financial features, health:** no.
6. **Store settings:** category Video Players & Editors; contact website
   `https://xmhrnx.github.io/goencode/`.
7. Promote the release to Production when the internal test works.

**Play notes.**

- `target_sdk = 36`: Google Play requires Android 16 (API 36) for new apps and
  updates since 31 Aug 2026, and the CLI's default, 35, is rejected. Raise it
  each August when Play's requirement moves.
- Google Play requires native code that loads on 16 KB page devices.
  `android/build-ffmpeg.sh` links ffmpeg with 16 KB alignment, and CI always
  builds it; don't release with an `ffmpeg.aar` from `fetchffmpeg`.
- Android 15 and later give "data sync" foreground services 6 hours a day. A
  batch running longer in the background is stopped by the system; keeping
  GoEncode open avoids it.
- The bundled ffmpeg includes x264 and is GPL v3, so the Android app is too.
  Keep the source available: this repository, plus ffmpeg 7.1.1 and x264 as
  fetched by `android/build-ffmpeg.sh`. Mention it in each release's notes.

## macOS (.dmg on GitHub)

1. Join the Apple Developer Program.
2. Create a "Developer ID Application" certificate, export it with its key as a
   `.p12`, and add it as `MACOS_CERT_BASE64` and `MACOS_CERT_PASS`.
3. Create an app-specific password at account.apple.com and add
   `GIODE_NOTARY_ID` (your Apple ID), `GIODE_NOTARY_TEAM_ID` and
   `GIODE_NOTARY_PASS`.
4. Publish a release. The `.dmg` is signed, notarized and stapled.

The app isn't sandboxed, so it runs Homebrew's ffmpeg (`/opt/homebrew/bin`,
`/usr/local/bin`) and MacPorts'. macOS asks once before GoEncode works in
Documents, Desktop, Downloads or external drives (`[macos.plist]` has the
texts). An ffmpeg downloaded as a zip from a website is quarantined and macOS
may refuse to run it; Homebrew's isn't.

## Store listing

Use the same text in both stores.

**Name:** GoEncode

**Short description** (Play, 80 characters at most; 72 here):

> Batch remux, encode and transcode video and audio with ffmpeg, on device

**Full description:**

> GoEncode is a front end for ffmpeg. Add any number of video and audio files,
> choose one set of settings, and remux, encode or transcode the whole batch.
>
> • One-click presets: Universal MP4, HEVC, AV1, GPU, WebM, remux, ProRes, GIF,
>   MP3, Opus and FLAC, with a quality slider and a resolution choice.
> • Copy streams untouched to change the container in seconds without losing
>   quality.
> • Every encoder your ffmpeg has. x264, x265, SVT-AV1, VP9, ProRes, AAC, Opus,
>   FLAC and more have curated settings, each explained in a popover.
> • Only valid choices: GoEncode shows the codecs and formats that fit the
>   container you pick, and fixes settings that no longer apply.
> • Preview a few seconds before encoding everything: speed, projected size and
>   the same frame from the source and the result side by side.
> • Hardware encoders (NVIDIA NVENC, Intel Quick Sync, AMD AMF, Media Foundation,
>   Android MediaCodec and others) are test-encoded first, so only working ones
>   are offered.
> • Pick audio and subtitle tracks by language or one by one, burn in
>   subtitles, keep chapters and metadata, and name outputs your way.
> • Blu-ray and DVD folders are added as their main title.
> • Encode several files at once, follow progress and get notified when the
>   batch is done.
>
> Everything happens on your device. GoEncode has no account, no ads, no
> tracking, and it doesn't use the internet.

Add for the **Microsoft Store**:

> Requires ffmpeg, which isn't included: install it with
> `winget install Gyan.FFmpeg`, or point GoEncode at ffmpeg.exe in Preferences.
> GoEncode shows how when it can't find it.

Add for **Google Play**:

> ffmpeg is built in. The app includes ffmpeg with x264 and is licensed under
> the GPL; the source is at github.com/xmhrnx/goencode.

**Search terms** (Microsoft Store, up to 7): ffmpeg, video converter, encoder,
transcode, remux, HEVC, AV1

**Category:** Microsoft Store: Multimedia design (or Photo & video).
Google Play: Video Players & Editors.

**Age rating:** everyone. No user content, communication, purchases or ads.

**Export compliance:** GoEncode contains no encryption.

## Screenshots

There are none in the repository yet. Take them from the running app
(`GIODE_MCP=1` serves screenshots; see the README), in dark mode, with real
files in the batch:

1. The main window with a batch running: files with progress, the simple
   settings and a preset chosen.
2. Advanced mode on an encoder's options, with an ⓘ popover open.
3. The preview: source and result side by side, with speed and projected size.
4. A file's tracks, with audio and subtitle tracks switched on and off.

The Microsoft Store wants at least one desktop screenshot (1366x768 or larger;
the window opens at 1280x820, so maximize it). Google Play wants 2 to 8 phone
screenshots (the phone layout: the Files and Settings tabs, and a running
batch) and a 1024x500 feature graphic.
