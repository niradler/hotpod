# HotPod

## Overview

Process Manager is a lightweight Go-based process management tool that allows you to start, stop, and manage long-running commands with support for keep-alive functionality. It is designed to run shell commands in a controlled manner and ensure reliability for daemon-like processes.

## Features

- Start and stop processes safely
- Prevent multiple instances of the same process from running
- Graceful shutdown with `SIGTERM`, fallback to `SIGKILL`
- Keep-alive functionality to keep the processes running
- Thread-safe implementation using Go’s sync mechanisms

## Usage

Download the binary from release page.
Run a command using the process manager:

```sh
./hotpod -command "python -m http.server 8000" -shell "sh" -keepalive=true
```

## API

### `/start`

Starts a new process with the specified command.

**Request:**

- Method: `POST`
- URL: `/start`
- Body (JSON):
  
  ```json
  {
    "command": "your-command-here",
    "replace": true
  }
  ```

  - `command`: The command to run.
  - `replace`: If `true`, replaces the currently running process.

**Response:**

- Status: `200 OK` if the process starts successfully.
- Status: `400 Bad Request` if the request body is invalid.
- Status: `500 Internal Server Error` if there is an error starting the process.

### `/stop`

Stops the currently running process.

**Request:**

- Method: `POST`
- URL: `/stop`

**Response:**

- Status: `200 OK` if the process stops successfully.
- Status: `500 Internal Server Error` if there is an error stopping the process.

## Configuration Options

| Option        | Description                                    | Default |
|--------------|-------------------------------------------------|---------|
| `-command`   | The command to run                              | `""`    |
| `-shell`     | The shell to execute the command (e.g., `sh`)   | `sh`    |
| `-keepalive` | Keep the processes running                      | `false` |

## Contributing

We welcome contributions from the community! To contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature-name`)
3. Commit your changes (`git commit -m 'Add new feature'`)
4. Push to your branch (`git push origin feature-name`)
5. Create a Pull Request

## License

This project is licensed under the [MIT License](LICENSE).

## Community & Support

If you encounter any issues or have feature requests, feel free to open an issue on GitHub.

Happy coding! 🚀
