package handler

import "net/http"

func RouteWebDevTools() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}

func RouteStaticFiles() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", StaticFileHandler)
	mux.HandleFunc("/css/", StaticFileHandler)
	mux.HandleFunc("/font-awesome/", StaticFileHandler)
	mux.HandleFunc("/js/", StaticFileHandler)
	mux.HandleFunc("/img/", StaticFileHandler)
	return mux
}

func RouteInstall() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", InstallWebHandler)
	return mux
}
