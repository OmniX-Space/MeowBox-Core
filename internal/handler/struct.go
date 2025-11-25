package handler

import "net/http"

// NotFoundMiddleware Middleware to handle 404 errors
type notFoundMiddleware struct {
	handler http.Handler
}

// ErrorPageData Data model for error page template
type errorPageData struct {
	StatusCode int
	Title      string
	Message    string
}

// InstallPageData Data model for install page template
type installPageData struct {
	StatusCode int
}

// IndexPageData Data model for index page template
type indexPageData struct {
	StatusCode int
	Title      string
	I18n       string
}
