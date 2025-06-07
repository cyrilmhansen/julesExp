# Go Mobile OpenGL Example

This application demonstrates basic OpenGL rendering and touch input using the Go Mobile (`golang.org/x/mobile`) libraries. It renders a green triangle on a red background. The triangle's color pulsates, and its position can be controlled by touch. An FPS (Frames Per Second) counter is also displayed.

## Desktop Execution

To run this application on your desktop (Linux or macOS):

1. Ensure you have Go installed.
2. Install the example dependencies:
   ```bash
   go get -d golang.org/x/mobile/example/basic
   ```
3. Run the application:
   ```bash
   go run ubermobile/main.go
   ```
   (Note: The original comments mentioned `go install golang.org/x/mobile/example/basic && basic`. However, since the code is now local in the `ubermobile` directory, `go run ubermobile/main.go` is more direct for this specific project structure. If it were to be installed globally, the original command would be relevant.)

## Android Build and Installation

To build and install this application as an Android APK:

1. Install the `gomobile` tool:
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```
2. Get the example dependencies (if not already done for desktop):
   ```bash
   go get -d golang.org/x/mobile/example/basic
   ```
3. Build the APK:
   ```bash
   gomobile build -o app.apk ./ubermobile
   ```
   (This will create `app.apk` in the current directory. The path `./ubermobile` assumes you are running the command from the root of this repository).
4. Install on a connected Android device or emulator:
   ```bash
   gomobile install ./ubermobile
   ```

This project is based on the `golang.org/x/mobile/example/basic` example.
