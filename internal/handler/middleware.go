package handler

import (
	"fmt"
	"net/http"

	"github.com/OmniX-Space/MeowBox-Core/internal/service"
)

func InjectWebServerHeaders(config *service.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Server.ShowServerVersion {
			version := service.GetVersion()
			serverHeader := fmt.Sprintf("MeowBox/%s", version)
			w.Header().Set("Server", serverHeader)
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Server", "MeowBox")
		next.ServeHTTP(w, r)
	})
}
