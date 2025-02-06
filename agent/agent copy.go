package main

// import (
// 	"bufio"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"flag"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/exec"
// 	"os/signal"
// 	"runtime"
// 	"sync"
// 	"syscall"
// 	"time"
// )

// type ProcessManager struct {
// 	Process      *exec.Cmd
// 	KeepAlive    bool
// 	ForceRestart bool
// 	Command      string
// 	Shell        string
// 	Mutex        sync.Mutex
// 	ProcessDone  chan struct{}
// 	LogStreamers []chan string
// 	LogMutex     sync.Mutex
// }

// // NewProcessManager initializes a new ProcessManager with the given configuration.
// func NewProcessManager(shell string, debug bool, forceRestart bool, command string) *ProcessManager {
// 	return &ProcessManager{
// 		KeepAlive:    debug,
// 		ForceRestart: forceRestart,
// 		Shell:        shell,
// 		Command:      command,
// 		ProcessDone:  make(chan struct{}),
// 		LogStreamers: []chan string{},
// 	}
// }

// func (pm *ProcessManager) StartProcess(command string, forceRestart bool) error {
// 	pm.Mutex.Lock()
// 	defer pm.Mutex.Unlock()

// 	if pm.Process != nil && forceRestart {
// 		pm.StopProcess()
// 	}

// 	if pm.Process != nil {
// 		return errors.New("process already running")
// 	}

// 	pm.Command = command

// 	var cmd *exec.Cmd
// 	if runtime.GOOS == "windows" {
// 		cmd = exec.Command(pm.Shell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", command)
// 	} else {
// 		cmd = exec.Command(pm.Shell, "-c", command)
// 	}

// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return fmt.Errorf("failed to capture stdout: %v", err)
// 	}
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		return fmt.Errorf("failed to capture stderr: %v", err)
// 	}

// 	if err := cmd.Start(); err != nil {
// 		return fmt.Errorf("failed to start process: %v", err)
// 	}

// 	pm.Process = cmd

// 	go pm.streamLogs(stdout)
// 	go pm.streamLogs(stderr)
// 	go pm.monitorProcess()

// 	return nil
// }

// func (pm *ProcessManager) StopProcess() error {
// 	pm.Mutex.Lock()
// 	defer pm.Mutex.Unlock()

// 	if pm.Process == nil {
// 		return errors.New("no process running")
// 	}

// 	if err := pm.Process.Process.Signal(syscall.SIGTERM); err != nil {
// 		fmt.Printf("failed to stop process: %v", err)
// 		if err := pm.Process.Process.Kill(); err != nil {
// 			return fmt.Errorf("failed to kill process: %v", err)
// 		}
// 	}

// 	pm.Process = nil
// 	return nil
// }

// func (pm *ProcessManager) monitorProcess() {
// 	err := pm.Process.Wait()

// 	pm.Mutex.Lock()
// 	select {
// 	case <-pm.ProcessDone:
// 		// Channel already closed
// 	default:
// 		close(pm.ProcessDone)
// 	}
// 	pm.Process = nil
// 	pm.Mutex.Unlock()

// 	pm.Mutex.Lock()
// 	pm.Process = nil
// 	pm.Mutex.Unlock()

// 	log.Printf("Process exited with error: %v\n", err)
// }

// func (pm *ProcessManager) streamLogs(pipe io.ReadCloser) {
// 	defer pipe.Close()
// 	scanner := bufio.NewScanner(pipe)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		fmt.Println(line)
// 		pm.LogMutex.Lock()
// 		for _, streamer := range pm.LogStreamers {
// 			streamer <- line
// 		}
// 		pm.LogMutex.Unlock()
// 	}
// }

// func main() {
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}

// 	defaultShell := "sh"
// 	if runtime.GOOS == "windows" {
// 		defaultShell = "powershell"
// 	}
// 	if userShell := os.Getenv("SHELL"); userShell != "" {
// 		defaultShell = userShell
// 	}

// 	serverPort := flag.String("port", port, "server port")
// 	shell := flag.String("shell", defaultShell, "Shell to use for executing commands")
// 	debug := flag.Bool("debug", false, "Enable debug mode")
// 	restart := flag.Bool("restart", true, "Force restart process")
// 	command := flag.String("command", "", "Command to start initially")
// 	host := flag.String("host", "", "host to listen on")
// 	flag.Parse()

// 	// Initialize ProcessManager with all flags
// 	pm := NewProcessManager(*shell, *debug, *restart, *command)

// 	if *command != "" {
// 		log.Println("Starting initial command...")
// 		if err := pm.StartProcess(*command, *restart); err != nil {
// 			log.Fatalf("Failed to start initial process: %v", err)
// 		}
// 	}

// 	server := &http.Server{Addr: *host + ":" + *serverPort}

// 	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
// 		var request struct {
// 			Command      string `json:"command"`
// 			ForceRestart bool   `json:"force_restart"`
// 		}
// 		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
// 			http.Error(w, "Invalid request body", http.StatusBadRequest)
// 			return
// 		}

// 		if err := pm.StartProcess(request.Command, request.ForceRestart); err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		fmt.Fprintf(w, "Process started successfully\n")
// 	})

// 	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
// 		if err := pm.StopProcess(); err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		fmt.Fprintf(w, "Process stopped successfully\n")
// 	})

// 	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
// 		flusher, ok := w.(http.Flusher)
// 		if !ok {
// 			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
// 			return
// 		}
// 		w.Header().Set("Content-Type", "text/event-stream")
// 		w.Header().Set("Cache-Control", "no-cache")
// 		w.Header().Set("Connection", "keep-alive")
// 		ch := make(chan string)
// 		pm.LogMutex.Lock()
// 		pm.LogStreamers = append(pm.LogStreamers, ch)
// 		pm.LogMutex.Unlock()
// 		defer func() {
// 			pm.LogMutex.Lock()
// 			for i, streamer := range pm.LogStreamers {
// 				if streamer == ch {
// 					pm.LogStreamers = append(pm.LogStreamers[:i], pm.LogStreamers[i+1:]...)
// 					break
// 				}
// 			}
// 			pm.LogMutex.Unlock()
// 			close(ch)
// 		}()
// 		for msg := range ch {
// 			fmt.Fprintf(w, "data: %s\n\n", msg)
// 			flusher.Flush()
// 		}
// 	})

// 	go func() {
// 		log.Printf("Starting server on :%s", port)
// 		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			log.Fatalf("Failed to start server: %v", err)
// 		}
// 	}()

// 	// Graceful shutdown
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit
// 	log.Println("Shutting down server...")

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	if err := server.Shutdown(ctx); err != nil {
// 		log.Fatalf("Server forced to shutdown: %v", err)
// 	}

// 	if pm.Process != nil {
// 		if err := pm.StopProcess(); err != nil {
// 			log.Printf("Failed to stop process: %v", err)
// 		}
// 	}

// 	log.Println("Server exited gracefully")
// }
