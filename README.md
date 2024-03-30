# gio-test

This is an example application written in Go using Gio UI. It performs DNS resolution and displays the result.

<img width="210" alt="Screenshot 2024-03-28 at 9 41 02 PM" src="https://github.com/amircybersec/gio-test/assets/117060873/a66ae9a3-6b4f-4e97-95f2-6a025866767d">
<img width="210" alt="Screenshot 2024-03-29 at 1 20 57 PM" src="https://github.com/amircybersec/gio-test/assets/117060873/a9d449e4-8fce-405d-a73d-71f2f9497325">


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

