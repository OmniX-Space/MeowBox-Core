package service

import (
	"log"
	"net/http"
)

func WebService() {
	log.Printf("[Info] Starting web service")
	log.Printf("[Info] Web service started on 0.0.0.0:2233")
	err := http.ListenAndServe("0.0.0.0:2233", nil)
	if err != nil {
		log.Fatalf("[Error] Failed to start server: %v", err)
	}
}
