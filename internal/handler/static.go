package handler

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/OmniX-Space/MeowBox-Core/internal/service"
)

//go:embed web
var webFiles embed.FS

//go:embed web/install.html
var installTemplateContent string

//go:embed web/index.html
var pageTemplateContent string

//go:embed web/httperr.html
var errorTemplateContent string

var (
	installTemplate *template.Template
	pageTemplate    *template.Template
	errorTemplate   *template.Template
	installOnce     sync.Once
	pageOnce        sync.Once
	errorOnce       sync.Once
)

// loadInstallTemplate Initialize template (executed only once)
func loadInstallTemplate() {
	installOnce.Do(func() {
		var err error
		installTemplate, err = template.New("install").Parse(installTemplateContent)
		if err != nil {
			log.Fatalf("[Error] Failed to parse install template: %v", err)
		}
	})
}

// loadPageTemplate Initialize template (executed only once)
func loadPageTemplate() {
	pageOnce.Do(func() {
		var err error
		pageTemplate, err = template.New("page").Parse(pageTemplateContent)
		if err != nil {
			log.Fatalf("[Error] Failed to parse page template: %v", err)
		}
	})
}

// loadErrorTemplate Initialize template (executed only once)
func loadErrorTemplate() {
	errorOnce.Do(func() {
		var err error
		errorTemplate, err = template.New("error").Parse(errorTemplateContent)
		if err != nil {
			log.Fatalf("[Error] Failed to parse error template: %v", err)
		}
	})
}

// StaticFileHandler Embedded static file service
func StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Clean path, make sure it starts with "web/"
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	// Concatenate embed.FS path
	filePath := "web/" + path
	data, err := webFiles.ReadFile(filePath)
	if err != nil {
		log.Printf("[Error] Static file not found: %s", filePath)
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	// Set Content-Type
	contentType := service.GetContentType(path)
	SetHeaders(w, contentType)

	_, _ = w.Write(data)
}

func SetHeaders(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}
