# Hotpod

A cross-platform process manager written in Go that allows you to start, stop, and monitor processes. It also provides an HTTP API for remote management and log streaming.

---

## Features

- **Cross-Platform**: Works on Windows, Linux, and macOS.
- **Process Management**:
  - Start and stop processes.
  - Restart processes automatically (keep-alive).
  - Force restart running processes.
- **Log Streaming**: Stream process logs in real-time via HTTP.
- **HTTP API**: Manage processes remotely using RESTful endpoints.
- **Customizable**: Configure the shell, command, and debug mode via flags or environment variables.

---

## Installation

### Prerequisites

- Go 1.16 or higher.
- A shell (e.g., `bash`, `sh`, or `powershell`).

## Usage

### Command-Line Flags

| Flag       | Description                         | Default Value                         |
| ---------- | ----------------------------------- | ------------------------------------- |
| `-shell`   | Shell to use for executing commands | `sh` (Unix) or `powershell` (Windows) |
| `-debug`   | Enable debug mode                   | `false`                               |
| `-command` | Command to start initially          | `""` (empty)                          |

Example:
```bash
./process-manager -shell powershell -command "Get-Process" -debug true
```

### Environment Variables

| Variable | Description                         | Default Value                         |
| -------- | ----------------------------------- | ------------------------------------- |
| `PORT`   | Port for the HTTP server            | `8080`                                |
| `SHELL`  | Shell to use for executing commands | `sh` (Unix) or `powershell` (Windows) |

Example:
```bash
export PORT=8081
export SHELL=bash
./process-manager
```

---

## HTTP API

The process manager exposes the following HTTP endpoints:

### Start a Process
- **Endpoint**: `POST /start`
- **Request Body**:
  ```json
  {
    "command": "string",      // Command to execute
    "force_restart": boolean  // Force restart if process is already running
  }
  ```
- **Response**: `200 OK` if successful.

Example:
```bash
curl -X POST http://localhost:8080/start -H "Content-Type: application/json" -d "{\"command\": \"Get-Process\", \"force_restart\": false}"
```

### Stop a Process
- **Endpoint**: `POST /stop`
- **Response**: `200 OK` if successful.

Example:
```bash
curl -X POST http://localhost:8080/stop
```

### Stream Logs
- **Endpoint**: `GET /logs`
- **Response**: A stream of process logs in Server-Sent Events (SSE) format.

Example:
```bash
curl http://localhost:8080/logs
```

## Example Use Cases

1. **Run a Long-Running Process**:
   ```bash
   ./process-manager -shell bash -command "while true; do echo 'Running'; sleep 5; done"
   ```

2. **Monitor System Processes**:
   ```bash
   ./process-manager -shell powershell -command "Get-Process"
   ```

3. **Stream Logs to a Dashboard**:
   Use the `/logs` endpoint to stream logs to a web-based dashboard.

---

## Configuration

### Shells
- **Unix-like Systems**: Use `sh`, `bash`, or any compatible shell.
- **Windows**: Use `powershell` or `cmd`.

### Commands
- Ensure the commands are compatible with the selected shell.
- For example:
  - Unix: `ls -l`, `ps aux`
  - Windows: `Get-Process`, `dir`

---

## Troubleshooting

### Issue: Command Not Found
- **Cause**: The command is not recognized by the shell.
- **Fix**: Use shell-compatible commands (e.g., `Get-Process` on Windows, `ps` on Unix).

### Issue: Process Not Terminating
- **Cause**: `syscall.SIGTERM` may not work on Windows.
- **Fix**: Use `cmd.Process.Kill()` for forceful termination.

### Issue: Logs Not Streaming
- **Cause**: The client may not support Server-Sent Events (SSE).
- **Fix**: Use a compatible client (e.g., `curl` or a modern browser).

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- Inspired by the need for a simple, cross-platform process manager.
- Built with Go's powerful standard library.

---

## Contact

For questions or feedback, please open an issue on GitHub or contact the maintainer.

---

Enjoy managing your processes with ease! 🚀