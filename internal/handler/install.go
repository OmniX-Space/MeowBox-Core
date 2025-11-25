package handler

import (
	"log"
	"net/http"

	"github.com/OmniX-Space/MeowBox-Core/internal/service"
)

func CheckInstall(config *service.Config) {
	if GetInstallLock(config) == false {
		log.Println("[Info] MeowBox-Core is not installed.")
		log.Println("[Info] Starting installation process.")
		server := service.CreateWebService(config)
		router := http.NewServeMux()
		router.Handle("/", RouteInstall())
		staticRouter := RouteStaticFiles()
		router.Handle("/favicon.ico", staticRouter)
		router.Handle("/css/", staticRouter)
		router.Handle("/font-awesome/", staticRouter)
		router.Handle("/js/", staticRouter)
		router.Handle("/img/", staticRouter)
		router.Handle("/.well-known/", RouteWebDevTools())
		server.Handler = InjectWebServerHeaders(config, router)
		service.ListenWebService(config, server)
		return
	}
	log.Println("[Info] MeowBox-Core is already installed.")
}

func InstallWebHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	data := installPageData{
		StatusCode: 200,
	}
	loadInstallTemplate()
	w.WriteHeader(200)
	SetHeaders(w, "text/html; charset=utf-8")
	if err := installTemplate.Execute(w, data); err != nil {
		log.Printf("[Error] Failed to render install page: %v", err)
	}
}

func GetInstallLock(config *service.Config) bool {
	db, err := service.ConnectDatabase(config.Database.Driver, config.Database.Host, config.Database.Port, config.Database.Username, config.Database.Password, config.Database.Database)
	if err != nil {
		log.Fatalf("[Error] Failed to connect to database: %s", err)
		return false
	}
	defer service.CloseDatabase(db)
	if service.TableExists(db, config.Database.Prefix, "system_config") {
		return true
	}
	return false
}

func InstallDatabase() error {
	return nil
}
