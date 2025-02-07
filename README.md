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

Download the binary from releases page.

[https://github.com/niradler/hotpod/releases](https://github.com/niradler/hotpod/releases)

Run a command using the process manager:

```sh
./hotpod -command "python -m http.server 8000" -shell "sh" -keepalive=true -host localhost -port 8080
```

[Example docker file](https://github.com/niradler/hotpod/blob/master/Dockerfile)

## API

[Swagger/Openapi](https://github.com/niradler/hotpod/blob/master/swagger.yaml)

## Configuration Options

| Option        | Description                                    | Default |
|--------------|-------------------------------------------------|---------|
| `-command`   | The command to run                              | `""`    |
| `-shell`     | The shell to execute the command (e.g., `sh`)   | `sh`    |
| `-keepalive` | Keep the processes running                      | `false` |
| `-host`      | The host to listen on                           | `localhost` |
| `-port`      | The server port                                 | `8080` |

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
