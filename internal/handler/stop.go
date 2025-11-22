package handler

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// HandleStop handles the stop signal and exits the program.
func HandleStop() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("[Info] Received signal: %v，stopping...\r\n", sig)
		os.Exit(0)
	}()
}
