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

To install with command line run:

```sh
adb install gio-test.apk
```

## iOS

Requirements:
- XCODE
- [ideviceinstaller](https://github.com/libimobiledevice/ideviceinstaller) (to install with command line on physical device)

Create `.app` file to install on simulator:

```sh
go run gioui.org/cmd/gogio -o gio.app -target ios -appid com.example.appleid .
```

replace `com.example.appleid` with your own app id. Make sure you have a valid certitificate and provisioning profile setup first.

Then run the following command to run the app on iPhone simulator:

```sh
xcrun simctl install booted gio-test.app
```

To install on physical device you must create `.ipa` file first

```sh
go run gioui.org/cmd/gogio -target ios -appid com.example.appleid .
```

Then run:

```sh
ideviceinstaller -i gio-test.ipa
```

