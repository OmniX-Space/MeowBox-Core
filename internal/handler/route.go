package handler

import "net/http"

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
