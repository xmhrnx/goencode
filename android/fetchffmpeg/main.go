// Command fetchffmpeg writes android/ffmpeg.aar from the ffmpeg and
// ffprobe in the APK of GoEncode's newest release that has one, for
// building the Android app where build-ffmpeg.sh can't run (Windows) or
// would take half an hour.
//
//	go run ./android/fetchffmpeg [-apk path-or-url] [-force]
//
// The APK carries the same lib/<abi>/libffmpeg.so and libffprobe.so the
// .aar puts under jni/<abi>, so they're copied across as they are.
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const releases = "https://api.github.com/repos/xmhrnx/goencode/releases"

const manifest = `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.goencode.ffmpeg" />
`

func main() {
	apk := flag.String("apk", "", "the APK to take ffmpeg from, a file or a URL (default: the newest release's)")
	out := flag.String("o", filepath.Join("android", "ffmpeg.aar"), "the .aar to write")
	force := flag.Bool("force", false, "replace an existing .aar")
	flag.Parse()

	if _, err := os.Stat(*out); err == nil && !*force {
		fmt.Printf("fetchffmpeg: %s exists (-force replaces it)\n", *out)
		return
	}
	if err := run(*apk, *out); err != nil {
		fmt.Fprintln(os.Stderr, "fetchffmpeg:", err)
		os.Exit(1)
	}
}

func run(apk, out string) error {
	if apk == "" {
		var err error
		if apk, err = releasedAPK(); err != nil {
			return err
		}
	}
	path := apk
	if strings.HasPrefix(apk, "https://") || strings.HasPrefix(apk, "http://") {
		tmp, err := download(apk)
		if err != nil {
			return err
		}
		defer os.Remove(tmp)
		path = tmp
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("%s isn't an APK: %w", apk, err)
	}
	defer zr.Close()

	f, err := os.Create(out + ".tmp")
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("AndroidManifest.xml")
	if err == nil {
		_, err = io.WriteString(w, manifest)
	}
	var abis []string
	for _, e := range zr.File {
		// lib/<abi>/libffmpeg.so, lib/<abi>/libffprobe.so
		parts := strings.Split(e.Name, "/")
		if err != nil || len(parts) != 3 || parts[0] != "lib" || (parts[2] != "libffmpeg.so" && parts[2] != "libffprobe.so") {
			continue
		}
		err = copyEntry(zw, e, "jni/"+parts[1]+"/"+parts[2])
		if parts[2] == "libffmpeg.so" {
			abis = append(abis, parts[1])
		}
	}
	if err == nil {
		err = zw.Close()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && len(abis) == 0 {
		err = fmt.Errorf("%s has no lib/<abi>/libffmpeg.so", apk)
	}
	if err != nil {
		os.Remove(out + ".tmp")
		return err
	}
	if err := os.Rename(out+".tmp", out); err != nil {
		return err
	}
	fmt.Printf("fetchffmpeg: wrote %s with ffmpeg for %s\n", out, strings.Join(abis, ", "))
	return nil
}

func copyEntry(zw *zip.Writer, e *zip.File, name string) error {
	r, err := e.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	return err
}

// releasedAPK finds the APK of the newest release that has one.
func releasedAPK() (string, error) {
	resp, err := http.Get(releases)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", releases, resp.Status)
	}
	var list []struct {
		Tag    string `json:"tag_name"`
		Draft  bool   `json:"draft"`
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return "", err
	}
	for _, r := range list {
		for _, a := range r.Assets {
			if !r.Draft && strings.HasSuffix(a.Name, ".apk") {
				fmt.Printf("fetchffmpeg: using the APK of release %s\n", r.Tag)
				return a.URL, nil
			}
		}
	}
	return "", fmt.Errorf("no release has an APK; pass one with -apk")
}

func download(url string) (string, error) {
	fmt.Printf("fetchffmpeg: downloading %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", url, resp.Status)
	}
	f, err := os.CreateTemp("", "goencode-*.apk")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), f.Close()
}
