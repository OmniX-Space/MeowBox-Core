package handler

import (
	"log"
	"net/http"

	"github.com/OmniX-Space/MeowBox-Core/internal/service"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {}

// ErrorHandler Common error response handler
func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int) {
	loadErrorTemplate()

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

	data := service.ErrorPageData{
		StatusCode: statusCode,
		Title:      title,
		Message:    message,
	}

	if err := errorTemplate.Execute(w, data); err != nil {
		log.Printf("[Error] Failed to render error page: %v", err)
	}
}
