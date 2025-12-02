package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const (
	TAG = "MeowEmbeddedMusicServer"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("[Warning] %s Loading .env file failed: %v\nUse the default configuration instead.\n", TAG, err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Printf("[Warning] %s PORT environment variable not set\nUse the default port 2233 instead.\n", TAG)
		port = "2233"
	}

	// Initialize user store
	InitUserStore()

	// Initialize playlist manager
	InitPlaylistManager()

	// Initialize device manager
	GetDeviceManager()

	// Register handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/stream_pcm", apiHandler)
	http.HandleFunc("/stream_live", streamLiveHandler) // 实时流式转码，边下载边播放
	http.HandleFunc("/api/search", HandleSearch) // Search API
	
	// Legacy favorite API (backward compatible)
	http.HandleFunc("/api/favorite/add", HandleAddToFavorite)
	http.HandleFunc("/api/favorite/remove", HandleRemoveFromFavorite)
	http.HandleFunc("/api/favorite/list", HandleGetFavorites)
	http.HandleFunc("/api/favorite/check", HandleCheckFavorite)
	
	// User authentication API
	http.HandleFunc("/api/auth/register", HandleRegister)
	http.HandleFunc("/api/auth/login", HandleLogin)
	http.HandleFunc("/api/auth/logout", HandleLogout)
	http.HandleFunc("/api/auth/me", HandleGetCurrentUser)
	
	// User playlist API (requires authentication)
	http.HandleFunc("/api/user/playlists", AuthMiddleware(HandleGetUserPlaylists))
	http.HandleFunc("/api/user/playlists/create", AuthMiddleware(HandleCreateUserPlaylist))
	http.HandleFunc("/api/user/playlists/add-song", AuthMiddleware(HandleAddSongToUserPlaylist))
	http.HandleFunc("/api/user/playlists/remove-song", AuthMiddleware(HandleRemoveSongFromUserPlaylist))
	http.HandleFunc("/api/user/playlists/delete", AuthMiddleware(HandleDeleteUserPlaylist))
	
	// ESP32 Device management API
	http.HandleFunc("/api/device/generate-code", AuthMiddleware(GenerateBindingCodeHandler)) // 生成绑定码（需要登录）
	http.HandleFunc("/api/device/bind-direct", DirectBindDeviceHandler)                       // Web端直接绑定设备（需要登录）
	http.HandleFunc("/api/device/list", ListUserDevicesHandler)                               // 列出用户设备（需要登录）
	http.HandleFunc("/api/device/unbind", UnbindDeviceHandler)                                // 解绑设备（需要登录）
	http.HandleFunc("/api/esp32/bind", BindDeviceHandler)                                     // ESP32绑定设备
	http.HandleFunc("/api/esp32/verify", VerifyDeviceHandler)                                 // 验证设备Token
	http.HandleFunc("/api/esp32/sync", SyncDeviceHandler)                                     // ESP32同步Token（用MAC地址）
	
	// Device binding page
	http.HandleFunc("/device-bind", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./theme/device-bind.html")
	})
	// 兼容用户常见的拼写错误
	http.HandleFunc("/device-bin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/device-bind", http.StatusMovedPermanently)
	})
	
	fmt.Printf("[Info] %s Started.\n喵波音律-音乐家园QQ交流群:865754861\n", TAG)
	fmt.Printf("[Info] Starting music server at port %s\n", port)

	// Create a channel to listen for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a server instance
	srv := &http.Server{
		Addr:              ":" + port,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,                // Disable the timeout for the response writer
		IdleTimeout:       30 * time.Minute, // Set the maximum duration for idle connections
		ReadHeaderTimeout: 10 * time.Second, // Limit the maximum duration for reading the headers of the request
		MaxHeaderBytes:    1 << 16,          // Limit the maximum request header size to 64KB
	}

	// Start the server
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println(err)
			sigChan <- syscall.SIGINT // Send a signal to shut down the server
		}
	}()

	// Create a channel to listen for standard input
	exitChan := make(chan struct{})

	go func() {
		for {
			var input string
			fmt.Scanln(&input)
			if input == "exit" {
				exitChan <- struct{}{}
				return
			}
		}
	}()

	// Monitor signals or exit signals from standard inputs
	select {
	case <-sigChan:
		fmt.Printf("[Info] Server is shutting down.\nGoodbye!\n")
	case <-exitChan:
		fmt.Printf("[Info] Server is shutting down.\nGoodbye!\n")
	}

	// Shut down the server
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Println(err)
	}
}
