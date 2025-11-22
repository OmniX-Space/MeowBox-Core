package core

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/OmniX-Space/MeowBox-Core/internal/handler"
	"github.com/OmniX-Space/MeowBox-Core/internal/service"
)

func Start() {
	Stop()
	log.Printf("[Info] Starting MeowBox Core...\r\n")
	config, err := service.GetConfig()
	if err != nil {
		log.Fatalf("[Error] Failed to load config: %v", err)
	}
	handler.CheckInstall(config)
}
func Stop() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("[Info] Received signal: %v，stopping...\r\n", sig)
		os.Exit(0)
	}()
}
