package platform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

// ErrNoOpener means files can't be opened in another app from here; use
// the share sheet instead.
var ErrNoOpener = errors.New("opening files isn't supported here")

var (
	publicOnce sync.Once
	publicDir  string
)

// PublicDir is Download/GoEncode in shared storage, where results go so
// the Files app shows them and other apps can open them. Android 11 and
// later let apps create files there by path; on older versions it's "",
// and results stay in the app.
func PublicDir() string {
	publicOnce.Do(func() {
		root := os.Getenv("EXTERNAL_STORAGE")
		if root == "" {
			root = "/storage/emulated/0"
		}
		// EXTERNAL_STORAGE is often /sdcard, a link; MediaStore records
		// the real path, which mediaID looks files up by.
		if real, err := filepath.EvalSymlinks(root); err == nil {
			root = real
		}
		dir := filepath.Join(root, "Download", "GoEncode")
		if os.MkdirAll(dir, 0o755) != nil {
			return
		}
		f, err := os.CreateTemp(dir, ".probe-*")
		if err != nil {
			return
		}
		f.Close()
		os.Remove(f.Name())
		publicDir = dir
	})
	return publicDir
}

// CanOpen reports whether Open works for the file at path: only files in
// shared storage, which other apps can read.
func CanOpen(path string) bool {
	dir := PublicDir()
	return dir != "" && strings.HasPrefix(path, dir+string(filepath.Separator))
}

// Open plays a result in the app the user picks for its type.
func Open(path string) error {
	if !CanOpen(path) {
		return ErrNoOpener
	}
	return startIntent(path, false)
}

// Share opens the share sheet for a result in shared storage. ok is
// false for other files, which the app's own share copies out first.
func Share(path string) (ok bool, err error) {
	if !CanOpen(path) {
		return false, nil
	}
	return true, startIntent(path, true)
}

const (
	flagGrantRead = 0x00000001 // Intent.FLAG_GRANT_READ_URI_PERMISSION
	flagNewTask   = 0x10000000 // Intent.FLAG_ACTIVITY_NEW_TASK
)

// startIntent shows path in a viewer (ACTION_VIEW) or the share sheet
// (ACTION_SEND), through its MediaStore content URI: Android doesn't let
// apps hand out file paths.
func startIntent(path string, share bool) (err error) {
	mime := MIMEType(path)
	ctx := jni.Object(app.AppContext())
	return jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) (err error) {
		// FindClass and Get*MethodID panic on a missing class or method.
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("android: %v", r)
			}
		}()
		uri, err := contentURI(env, ctx, path, mime)
		if err != nil {
			return err
		}
		intentCls := jni.FindClass(env, "android/content/Intent")
		action := "android.intent.action.VIEW"
		if share {
			action = "android.intent.action.SEND"
		}
		intent, err := jni.NewObject(env, intentCls, jni.GetMethodID(env, intentCls, "<init>", "(Ljava/lang/String;)V"),
			jni.Value(jni.JavaString(env, action)))
		if err != nil {
			return err
		}
		if share {
			if _, err := jni.CallObjectMethod(env, intent, jni.GetMethodID(env, intentCls, "setType", "(Ljava/lang/String;)Landroid/content/Intent;"),
				jni.Value(jni.JavaString(env, mime))); err != nil {
				return err
			}
			if _, err := jni.CallObjectMethod(env, intent, jni.GetMethodID(env, intentCls, "putExtra", "(Ljava/lang/String;Landroid/os/Parcelable;)Landroid/content/Intent;"),
				jni.Value(jni.JavaString(env, "android.intent.extra.STREAM")), jni.Value(uri)); err != nil {
				return err
			}
			// The grant reaches the chosen app through the clip data.
			clipCls := jni.FindClass(env, "android/content/ClipData")
			clip, err := jni.CallStaticObjectMethod(env, clipCls, jni.GetStaticMethodID(env, clipCls, "newRawUri", "(Ljava/lang/CharSequence;Landroid/net/Uri;)Landroid/content/ClipData;"),
				jni.Value(jni.JavaString(env, "")), jni.Value(uri))
			if err != nil {
				return err
			}
			if err := jni.CallVoidMethod(env, intent, jni.GetMethodID(env, intentCls, "setClipData", "(Landroid/content/ClipData;)V"), jni.Value(clip)); err != nil {
				return err
			}
		} else if _, err := jni.CallObjectMethod(env, intent, jni.GetMethodID(env, intentCls, "setDataAndType", "(Landroid/net/Uri;Ljava/lang/String;)Landroid/content/Intent;"),
			jni.Value(uri), jni.Value(jni.JavaString(env, mime))); err != nil {
			return err
		}
		addFlags := jni.GetMethodID(env, intentCls, "addFlags", "(I)Landroid/content/Intent;")
		if _, err := jni.CallObjectMethod(env, intent, addFlags, jni.Value(flagGrantRead|flagNewTask)); err != nil {
			return err
		}
		if share {
			intent, err = jni.CallStaticObjectMethod(env, intentCls, jni.GetStaticMethodID(env, intentCls, "createChooser", "(Landroid/content/Intent;Ljava/lang/CharSequence;)Landroid/content/Intent;"),
				jni.Value(intent), jni.Value(0))
			if err != nil {
				return err
			}
			if _, err := jni.CallObjectMethod(env, intent, addFlags, jni.Value(flagNewTask)); err != nil {
				return err
			}
		}
		ctxCls := jni.GetObjectClass(env, ctx)
		return jni.CallVoidMethod(env, ctx, jni.GetMethodID(env, ctxCls, "startActivity", "(Landroid/content/Intent;)V"), jni.Value(intent))
	})
}

// contentURI finds the MediaStore item for a file in shared storage,
// asking the media scanner for it when it isn't indexed yet. The link is
// to the Downloads collection, like giode's own share: it's the one this
// app may grant other apps; the video collection's link opened a sheet
// with no apps in it.
func contentURI(env jni.Env, ctx jni.Object, path, mime string) (jni.Object, error) {
	id, err := mediaID(env, ctx, path)
	if err != nil {
		return 0, err
	}
	if id < 0 {
		scanner := jni.FindClass(env, "android/media/MediaScannerConnection")
		strCls := jni.FindClass(env, "java/lang/String")
		paths := jni.NewObjectArray(env, 1, strCls, jni.Object(jni.JavaString(env, path)))
		mimes := jni.NewObjectArray(env, 1, strCls, jni.Object(jni.JavaString(env, mime)))
		if err := jni.CallStaticVoidMethod(env, scanner, jni.GetStaticMethodID(env, scanner, "scanFile", "(Landroid/content/Context;[Ljava/lang/String;[Ljava/lang/String;Landroid/media/MediaScannerConnection$OnScanCompletedListener;)V"),
			jni.Value(ctx), jni.Value(paths), jni.Value(mimes), jni.Value(0)); err != nil {
			return 0, err
		}
		for i := 0; i < 25 && id < 0; i++ {
			time.Sleep(200 * time.Millisecond)
			if id, err = mediaID(env, ctx, path); err != nil {
				return 0, err
			}
		}
		if id < 0 {
			return 0, fmt.Errorf("Android hasn't indexed %s yet", filepath.Base(path))
		}
	}
	uriCls := jni.FindClass(env, "android/net/Uri")
	return jni.CallStaticObjectMethod(env, uriCls, jni.GetStaticMethodID(env, uriCls, "parse", "(Ljava/lang/String;)Landroid/net/Uri;"),
		jni.Value(jni.JavaString(env, "content://media/external/downloads/"+strconv.FormatInt(id, 10))))
}

// mediaID returns the MediaStore ID of the file at path, or -1.
func mediaID(env jni.Env, ctx jni.Object, path string) (int64, error) {
	ctxCls := jni.GetObjectClass(env, ctx)
	resolver, err := jni.CallObjectMethod(env, ctx, jni.GetMethodID(env, ctxCls, "getContentResolver", "()Landroid/content/ContentResolver;"))
	if err != nil {
		return -1, err
	}
	filesCls := jni.FindClass(env, "android/provider/MediaStore$Files")
	files, err := jni.CallStaticObjectMethod(env, filesCls, jni.GetStaticMethodID(env, filesCls, "getContentUri", "(Ljava/lang/String;)Landroid/net/Uri;"),
		jni.Value(jni.JavaString(env, "external")))
	if err != nil {
		return -1, err
	}
	strCls := jni.FindClass(env, "java/lang/String")
	projection := jni.NewObjectArray(env, 1, strCls, jni.Object(jni.JavaString(env, "_id")))
	args := jni.NewObjectArray(env, 1, strCls, jni.Object(jni.JavaString(env, path)))
	resolverCls := jni.GetObjectClass(env, resolver)
	cursor, err := jni.CallObjectMethod(env, resolver, jni.GetMethodID(env, resolverCls, "query", "(Landroid/net/Uri;[Ljava/lang/String;Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;)Landroid/database/Cursor;"),
		jni.Value(files), jni.Value(projection), jni.Value(jni.JavaString(env, "_data=?")), jni.Value(args), jni.Value(0))
	if err != nil || cursor == 0 {
		return -1, err
	}
	cursorCls := jni.GetObjectClass(env, cursor)
	defer jni.CallVoidMethod(env, cursor, jni.GetMethodID(env, cursorCls, "close", "()V"))
	found, err := jni.CallBooleanMethod(env, cursor, jni.GetMethodID(env, cursorCls, "moveToFirst", "()Z"))
	if err != nil || !found {
		return -1, err
	}
	return jni.CallLongMethod(env, cursor, jni.GetMethodID(env, cursorCls, "getLong", "(I)J"), jni.Value(0))
}
