package handler

import (
	"log"
	"net/http"
)

// ErrorHandler Common error response handler
func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int) {
	loadErrorTemplate()
	log.Printf("[Info] [Web Access] Handler http error page: %d, on path: %s", statusCode, r.URL.Path)

	w.WriteHeader(statusCode)
	SetHeaders(w, "text/html; charset=utf-8")

	var title, message string
	switch statusCode {
	case http.StatusNotFound:
		title = ""
		message = ""
	case http.StatusInternalServerError:
		title = ""
		message = ""
	case http.StatusBadRequest:
		title = ""
		message = ""
	case http.StatusForbidden:
		title = ""
		message = ""
	default:
		title = "Error"
		message = "An unexpected error occurred."
	}

	data := errorPageData{
		StatusCode: statusCode,
		Title:      title,
		Message:    message,
	}

	if err := errorTemplate.Execute(w, data); err != nil {
		log.Printf("[Error] Failed to render error page: %v", err)
	}
}
