# gio-test



To run the app:

```sh
go run github.com/fortuna/gio-test@latest
```

From the repository root:

```sh
go run .
```

To package:

```sh
go run gioui.org/cmd/gogio -target macos .
```

## Android

Requirements:
- Android SDK, with Android 31+
- Android NDK

You need to set `ANDROID_SDK_ROOT` and `ANDROID_NDK_ROOT`.

Create APK:

```sh
go run gioui.org/cmd/gogio -target android .
$ANDROID_SDK_ROOT/platform-tools/adb install gio-test.apk
```
