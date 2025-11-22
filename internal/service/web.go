package service

import (
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
	"time"
)

func getServerAddress(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}

func CreateWebService() *http.Server {
	log.Printf("[Info] Create web service")
	config, err := GetConfig()
	if err != nil {
		log.Fatalf("[Error] Failed to load config: %v", err)
	}

	addr := getServerAddress(config.Server.Host, config.Server.Port)

	server := &http.Server{
		Addr:           addr,
		ReadTimeout:    time.Duration(config.Server.Advanced.Readtimeout) * time.Second,
		WriteTimeout:   time.Duration(config.Server.Advanced.Writetimeout) * time.Second,
		IdleTimeout:    time.Duration(config.Server.Advanced.Idletimeout) * time.Second,
		MaxHeaderBytes: config.Server.Advanced.Maxheaderbytes << 16,
	}
	if config.Server.Tls.Enabled {
		log.Printf("[Info] TLS enabled")
		server.TLSConfig = &tls.Config{
			PreferServerCipherSuites: true,
			SessionTicketsDisabled:   false,
		}
	}
	return server
}

func ListenWebService(server *http.Server) {
	config, err := GetConfig()
	if err != nil {
		log.Fatalf("[Error] Failed to load config: %v", err)
	}
	addr := getServerAddress(config.Server.Host, config.Server.Port)
	if config.Server.Tls.Enabled {
		log.Printf("[Info] Starting HTTPS server on %s", addr)
		err := server.ListenAndServeTLS(
			config.Server.Tls.Cert,
			config.Server.Tls.Key,
		)
		if err != nil {
			log.Fatalf("[Error] HTTPS server failed: %v", err)
		}
	} else {
		log.Printf("[Info] Starting HTTP server on %s", addr)
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("[Error] HTTP server failed: %v", err)
		}
	}
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("[Error] Web server terminated unexpectedly: %v", err)
	}
}
