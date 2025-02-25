package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/go-chi/chi/v5"
)

func main() {

	PrintLogo()

	port := os.Getenv("PORT")
	args := []string{"-c"}
	if port == "" {
		port = "8080"
	}

	defaultShell := "sh"
	if runtime.GOOS == "windows" {
		defaultShell = "powershell"
		args = []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-Command"}
	}
	if userShell := os.Getenv("SHELL"); userShell != "" {
		defaultShell = userShell
	}

	serverPort := flag.String("port", port, "server port")
	shell := flag.String("shell", defaultShell, "Shell to use for executing commands")
	keepAlive := flag.Bool("keepAlive", false, "Keep process alive after command exits")
	command := flag.String("command", "", "Command to start initially")
	host := flag.String("host", "localhost", "host to listen on")
	flag.Parse()

	log.Printf("Starting with configuration - Port: %s, Shell: %s, KeepAlive: %t, Command: %s, Host: %s\n", *serverPort, *shell, *keepAlive, *command, *host)

	pm := NewProcessManager(NewProcessOpt{
		Command:   *command,
		KeepAlive: *keepAlive,
		Shell:     *shell,
		Args:      args,
	})

	router := chi.NewRouter()

	router.Get("/_health", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		message := []byte("Process is running")

		if pm.Process == nil {
			status = http.StatusServiceUnavailable
			message = []byte("Process is not running")
		}

		if pm.Process != nil && pm.Process.ProcessState != nil && pm.Process.ProcessState.Exited() {
			status = http.StatusServiceUnavailable
			message = []byte("Process has exited")
		}

		if !pm.KeepAlive {
			w.WriteHeader(status)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Write(message)
	})

	router.Put("/process", func(w http.ResponseWriter, r *http.Request) {
		var options struct {
			KeepAlive bool     `json:"keepAlive"`
			Shell     string   `json:"shell"`
			Args      []string `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		pm.KeepAlive = options.KeepAlive
		if options.Shell != "" {
			pm.Shell = options.Shell
		}
		if len(options.Args) > 0 {
			pm.Args = args
		}

		w.WriteHeader(http.StatusOK)
		response := struct {
			KeepAlive bool     `json:"keepAlive"`
			Shell     string   `json:"shell"`
			Args      []string `json:"args"`
		}{
			KeepAlive: pm.KeepAlive,
			Shell:     pm.Shell,
			Args:      pm.Args,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

		return
	})

	router.Post("/start", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Command string   `json:"command"`
			Replace bool     `json:"replace"`
			Env     []string `json:"env"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := pm.StartProcess(request.Command, request.Replace, &request.Env); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Process started successfully"))
		return
	})

	router.Post("/stop", func(w http.ResponseWriter, r *http.Request) {
		if err := pm.StopProcess(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Process stopped successfully"))
		return
	})

	go func() {
		err := http.ListenAndServe(*host+":"+*serverPort, router)
		log.Fatalf("Failed to start server: %v", err)
		panic(err)
	}()

	go func() {
		if *command != "" {
			log.Println("Starting initial command...")
			if err := pm.StartProcess(*command, false, nil); err != nil {
				log.Fatalf("Failed to start initial process: %v", err)
			}
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if pm.Process != nil {
		if err := pm.StopProcess(); err != nil {
			log.Printf("Failed to stop process: %v", err)
		}
	}

	log.Println("Server exited gracefully")
	os.Exit(0)
}
