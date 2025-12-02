package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Playlist represents a user's playlist
type Playlist struct {
	Name  string      `json:"name"`
	Songs []MusicItem `json:"songs"`
}

// PlaylistManager manages all playlists
type PlaylistManager struct {
	mu            sync.RWMutex
	Playlists     map[string]*Playlist      // Legacy playlists (backward compatible)
	UserPlaylists map[string][]*UserPlaylist // key: user ID, value: user's playlists
	filePath      string
	userFilePath  string
}

var playlistManager *PlaylistManager

// InitPlaylistManager initializes the playlist manager
func InitPlaylistManager() {
	playlistManager = &PlaylistManager{
		Playlists:     make(map[string]*Playlist),
		UserPlaylists: make(map[string][]*UserPlaylist),
		filePath:      "./files/playlists.json",
		userFilePath:  "./files/user_playlists.json",
	}
	
	// Create files directory if it doesn't exist
	os.MkdirAll("./files", 0755)
	
	// Load existing playlists (backward compatible)
	playlistManager.loadFromFile()
	
	// Load user playlists
	playlistManager.loadUserPlaylists()
	
	// Initialize "我喜欢" playlist if it doesn't exist (backward compatible)
	if _, exists := playlistManager.Playlists["favorite"]; !exists {
		playlistManager.Playlists["favorite"] = &Playlist{
			Name:  "我喜欢",
			Songs: []MusicItem{},
		}
		playlistManager.saveToFile()
	}
}

// loadFromFile loads playlists from JSON file
func (pm *PlaylistManager) loadFromFile() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	return json.Unmarshal(data, &pm.Playlists)
}

// saveToFile saves playlists to JSON file
// NOTE: Caller must hold the lock!
func (pm *PlaylistManager) saveToFile() error {
	data, err := json.MarshalIndent(pm.Playlists, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.filePath, data, 0644)
}

// AddToPlaylist adds a song to a playlist
func (pm *PlaylistManager) AddToPlaylist(playlistName string, song MusicItem) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	playlist, exists := pm.Playlists[playlistName]
	if !exists {
		playlist = &Playlist{
			Name:  playlistName,
			Songs: []MusicItem{},
		}
		pm.Playlists[playlistName] = playlist
	}

	// Check if song already exists in playlist
	for _, s := range playlist.Songs {
		if s.Title == song.Title && s.Artist == song.Artist {
			return fmt.Errorf("song already exists in playlist")
		}
	}

	playlist.Songs = append(playlist.Songs, song)
	return pm.saveToFile()
}

// RemoveFromPlaylist removes a song from a playlist
func (pm *PlaylistManager) RemoveFromPlaylist(playlistName string, title, artist string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	playlist, exists := pm.Playlists[playlistName]
	if !exists {
		return fmt.Errorf("playlist not found")
	}

	for i, song := range playlist.Songs {
		if song.Title == title && song.Artist == artist {
			playlist.Songs = append(playlist.Songs[:i], playlist.Songs[i+1:]...)
			return pm.saveToFile()
		}
	}

	return fmt.Errorf("song not found in playlist")
}

// GetPlaylist returns a playlist
func (pm *PlaylistManager) GetPlaylist(playlistName string) (*Playlist, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	playlist, exists := pm.Playlists[playlistName]
	if !exists {
		return nil, fmt.Errorf("playlist not found")
	}

	return playlist, nil
}

// HTTP Handlers

// HandleAddToFavorite handles adding a song to favorites (user-specific)
func HandleAddToFavorite(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[API] Add to favorite request received from %s\n", r.RemoteAddr)
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var userID string
	
	// First, check for ESP32 device token (X-Device-Token header)
	deviceToken := r.Header.Get("X-Device-Token")
	if deviceToken != "" {
		dm := GetDeviceManager()
		device, err := dm.VerifyToken(deviceToken)
		if err == nil {
			userID = userStore.GetUserIDByUsername(device.Username)
			fmt.Printf("[API] ESP32 device adding favorite (user: %s, userID: %s)\n", device.Username, userID)
		}
	}
	
	// If not ESP32, check for web user token (Authorization header)
	if userID == "" {
		token := r.Header.Get("Authorization")
		if token != "" {
			token = strings.TrimPrefix(token, "Bearer ")
			user, err := userStore.GetUserByToken(token)
			if err == nil {
				userID = user.ID
			}
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[API] Error reading body: %v\n", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	fmt.Printf("[API] Received body: %s\n", string(body))

	var song MusicItem
	err = json.Unmarshal(body, &song)
	if err != nil {
		fmt.Printf("[API] Error parsing JSON: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("[API] Adding to favorites: %s - %s (userID: %s)\n", song.Artist, song.Title, userID)

	// If user is authenticated, add to user's "我喜欢" playlist
	if userID != "" {
		// Get user's "我喜欢" playlist (first playlist)
		playlists := playlistManager.GetUserPlaylists(userID)
		
		// If user has no playlists, initialize them first
		if len(playlists) == 0 {
			fmt.Printf("[API] User %s has no playlists, initializing...\n", userID)
			playlistManager.InitializeUserPlaylists(userID)
			playlists = playlistManager.GetUserPlaylists(userID)
		}
		
		if len(playlists) > 0 {
			favoritePlaylist := playlists[0] // "我喜欢" is always the first playlist
			err = playlistManager.AddSongToUserPlaylist(userID, favoritePlaylist.ID, song)
			if err != nil {
				if err.Error() == "song already exists in playlist" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "success",
						"message": "歌曲已在收藏列表中",
					})
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Playlist initialization failed
			fmt.Printf("[Error] Failed to initialize playlists for user %s\n", userID)
			http.Error(w, "Failed to initialize user playlists", http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to global favorite playlist for anonymous users
		err = playlistManager.AddToPlaylist("favorite", song)
		if err != nil {
			if err.Error() == "song already exists in playlist" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "success",
					"message": "歌曲已在收藏列表中",
				})
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Printf("[Info] Added to favorites: %s - %s\n", song.Artist, song.Title)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "收藏成功",
	})
}

// HandleRemoveFromFavorite handles removing a song from favorites (user-specific)
func HandleRemoveFromFavorite(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[API] Remove from favorite request received from %s\n", r.RemoteAddr)
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var userID string
	
	// First, check for ESP32 device token (X-Device-Token header)
	deviceToken := r.Header.Get("X-Device-Token")
	if deviceToken != "" {
		dm := GetDeviceManager()
		device, err := dm.VerifyToken(deviceToken)
		if err == nil {
			userID = userStore.GetUserIDByUsername(device.Username)
			fmt.Printf("[API] ESP32 device removing favorite (user: %s, userID: %s)\n", device.Username, userID)
		}
	}
	
	// If not ESP32, check for web user token (Authorization header)
	if userID == "" {
		token := r.Header.Get("Authorization")
		if token != "" {
			token = strings.TrimPrefix(token, "Bearer ")
			user, err := userStore.GetUserByToken(token)
			if err == nil {
				userID = user.ID
			}
		}
	}

	// Read song info from body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	var song MusicItem
	err = json.Unmarshal(body, &song)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	title := song.Title
	artist := song.Artist

	if title == "" || artist == "" {
		http.Error(w, "Missing title or artist", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("[API] Removing from favorites: %s - %s (userID: %s)\n", artist, title, userID)

	// If user is authenticated, remove from user's "我喜欢" playlist
	if userID != "" {
		playlists := playlistManager.GetUserPlaylists(userID)
		
		// If user has no playlists, nothing to remove
		if len(playlists) == 0 {
			fmt.Printf("[API] User %s has no playlists, nothing to remove\n", userID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "歌曲不在收藏列表中",
			})
			return
		}
		
		favoritePlaylist := playlists[0]
		err = playlistManager.RemoveSongFromUserPlaylist(userID, favoritePlaylist.ID, title, artist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to global favorite playlist
		err = playlistManager.RemoveFromPlaylist("favorite", title, artist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Printf("[Info] Removed from favorites: %s - %s\n", artist, title)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "取消收藏成功",
	})
}

// HandleGetFavorites handles getting the favorite playlist (user-specific)
func HandleGetFavorites(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[API] Get favorites request received from %s\n", r.RemoteAddr)
	
	var userID string
	
	// First, check for ESP32 device token (X-Device-Token header)
	deviceToken := r.Header.Get("X-Device-Token")
	if deviceToken != "" {
		dm := GetDeviceManager()
		device, err := dm.VerifyToken(deviceToken)
		if err == nil {
			// Device.Username is the username, need to convert to userID
			userID = userStore.GetUserIDByUsername(device.Username)
			fmt.Printf("[API] ESP32 device authenticated: %s (user: %s, userID: %s)\n", device.MAC, device.Username, userID)
		} else {
			fmt.Printf("[API] Invalid device token: %v\n", err)
		}
	}
	
	// If not ESP32, check for web user token (Authorization header)
	if userID == "" {
		token := r.Header.Get("Authorization")
		if token != "" {
			token = strings.TrimPrefix(token, "Bearer ")
			user, err := userStore.GetUserByToken(token)
			if err == nil {
				userID = user.ID
				fmt.Printf("[API] Web user authenticated: %s\n", userID)
			}
		}
	}

	// If user is authenticated, return user's "我喜欢" playlist
	if userID != "" {
		playlists := playlistManager.GetUserPlaylists(userID)
		
		// If user has no playlists, initialize them first
		if len(playlists) == 0 {
			fmt.Printf("[API] User %s has no playlists, initializing...\n", userID)
			playlistManager.InitializeUserPlaylists(userID)
			playlists = playlistManager.GetUserPlaylists(userID)
		}
		
		if len(playlists) > 0 {
			favoritePlaylist := playlists[0] // "我喜欢" is always the first playlist
			fmt.Printf("[API] Returning user %s favorites: %d songs\n", userID, len(favoritePlaylist.Songs))
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return just the songs array for frontend compatibility
			json.NewEncoder(w).Encode(favoritePlaylist.Songs)
			return
		}
		
		// If still no playlists (initialization failed), return empty array
		fmt.Printf("[API] User %s playlists initialization failed, returning empty\n", userID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]MusicItem{})
		return
	}
	
	// Only return global favorite playlist if user is NOT authenticated
	fmt.Printf("[API] No authentication found, returning global favorites\n")
	playlist, err := playlistManager.GetPlaylist("favorite")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]MusicItem{}) // Return empty array
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(playlist.Songs)
}

// HandleCheckFavorite checks if a song is in favorites
func HandleCheckFavorite(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	artist := r.URL.Query().Get("artist")

	if title == "" || artist == "" {
		http.Error(w, "Missing title or artist", http.StatusBadRequest)
		return
	}

	playlist, err := playlistManager.GetPlaylist("favorite")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"is_favorite": false})
		return
	}

	for _, song := range playlist.Songs {
		if song.Title == title && song.Artist == artist {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"is_favorite": true})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"is_favorite": false})
}

// User Playlist Management Functions

// loadUserPlaylists loads user playlists from JSON file
func (pm *PlaylistManager) loadUserPlaylists() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.userFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	return json.Unmarshal(data, &pm.UserPlaylists)
}

// saveUserPlaylists saves user playlists to JSON file
// NOTE: Caller must hold the lock!
func (pm *PlaylistManager) saveUserPlaylists() error {
	data, err := json.MarshalIndent(pm.UserPlaylists, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.userFilePath, data, 0644)
}

// InitializeUserPlaylists creates default playlists for a new user
func (pm *PlaylistManager) InitializeUserPlaylists(userID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.UserPlaylists[userID]; exists {
		return // Already initialized
	}

	// Create "我喜欢" playlist for the user
	favoritePlaylist := &UserPlaylist{
		ID:          generateID(),
		UserID:      userID,
		Name:        "我喜欢",
		Description: "我喜欢的音乐",
		Songs:       []MusicItem{},
		IsPublic:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	pm.UserPlaylists[userID] = []*UserPlaylist{favoritePlaylist}
	pm.saveUserPlaylists()
}

// CreateUserPlaylist creates a new playlist for a user
func (pm *PlaylistManager) CreateUserPlaylist(userID, name, description string) (*UserPlaylist, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	playlist := &UserPlaylist{
		ID:          generateID(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Songs:       []MusicItem{},
		IsPublic:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	pm.UserPlaylists[userID] = append(pm.UserPlaylists[userID], playlist)
	pm.saveUserPlaylists()

	return playlist, nil
}

// GetUserPlaylists returns all playlists for a user
func (pm *PlaylistManager) GetUserPlaylists(userID string) []*UserPlaylist {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.UserPlaylists[userID]
}

// GetUserPlaylistByID returns a specific playlist
func (pm *PlaylistManager) GetUserPlaylistByID(userID, playlistID string) (*UserPlaylist, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	playlists, exists := pm.UserPlaylists[userID]
	if !exists {
		return nil, fmt.Errorf("no playlists found for user")
	}

	for _, playlist := range playlists {
		if playlist.ID == playlistID {
			return playlist, nil
		}
	}

	return nil, fmt.Errorf("playlist not found")
}

// AddSongToUserPlaylist adds a song to a user's playlist
func (pm *PlaylistManager) AddSongToUserPlaylist(userID, playlistID string, song MusicItem) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	playlists, exists := pm.UserPlaylists[userID]
	if !exists {
		return fmt.Errorf("no playlists found for user")
	}

	for _, playlist := range playlists {
		if playlist.ID == playlistID {
			// Check if song already exists
			for _, s := range playlist.Songs {
				if s.Title == song.Title && s.Artist == song.Artist {
					return fmt.Errorf("song already exists in playlist")
				}
			}

			playlist.Songs = append(playlist.Songs, song)
			playlist.UpdatedAt = time.Now()
			return pm.saveUserPlaylists()
		}
	}

	return fmt.Errorf("playlist not found")
}

// RemoveSongFromUserPlaylist removes a song from a user's playlist
func (pm *PlaylistManager) RemoveSongFromUserPlaylist(userID, playlistID, title, artist string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	playlists, exists := pm.UserPlaylists[userID]
	if !exists {
		return fmt.Errorf("no playlists found for user")
	}

	for _, playlist := range playlists {
		if playlist.ID == playlistID {
			for i, song := range playlist.Songs {
				if song.Title == title && song.Artist == artist {
					playlist.Songs = append(playlist.Songs[:i], playlist.Songs[i+1:]...)
					playlist.UpdatedAt = time.Now()
					return pm.saveUserPlaylists()
				}
			}
			return fmt.Errorf("song not found in playlist")
		}
	}

	return fmt.Errorf("playlist not found")
}

// DeleteUserPlaylist deletes a user's playlist
func (pm *PlaylistManager) DeleteUserPlaylist(userID, playlistID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	playlists, exists := pm.UserPlaylists[userID]
	if !exists {
		return fmt.Errorf("no playlists found for user")
	}

	for i, playlist := range playlists {
		if playlist.ID == playlistID {
			// Don't allow deleting "我喜欢" playlist
			if playlist.Name == "我喜欢" && i == 0 {
				return fmt.Errorf("cannot delete favorite playlist")
			}

			pm.UserPlaylists[userID] = append(playlists[:i], playlists[i+1:]...)
			return pm.saveUserPlaylists()
		}
	}

	return fmt.Errorf("playlist not found")
}

// User Playlist HTTP Handlers

// HandleCreateUserPlaylist handles creating a new playlist
func HandleCreateUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Playlist name is required", http.StatusBadRequest)
		return
	}

	playlist, err := playlistManager.CreateUserPlaylist(userID, req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("[Info] User %s created playlist: %s\n", userID, req.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playlist)
}

// HandleGetUserPlaylists handles getting all user playlists
func HandleGetUserPlaylists(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlists := playlistManager.GetUserPlaylists(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlists)
}

// HandleAddSongToUserPlaylist handles adding a song to a playlist
func HandleAddSongToUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistID := r.URL.Query().Get("playlist_id")
	if playlistID == "" {
		http.Error(w, "Missing playlist_id parameter", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var song MusicItem
	if err := json.Unmarshal(body, &song); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = playlistManager.AddSongToUserPlaylist(userID, playlistID, song)
	if err != nil {
		if err.Error() == "song already exists in playlist" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "歌曲已在歌单中",
			})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("[Info] User %s added song to playlist %s: %s - %s\n", userID, playlistID, song.Artist, song.Title)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "添加成功",
	})
}

// HandleRemoveSongFromUserPlaylist handles removing a song from a playlist
func HandleRemoveSongFromUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistID := r.URL.Query().Get("playlist_id")
	title := r.URL.Query().Get("title")
	artist := r.URL.Query().Get("artist")

	if playlistID == "" || title == "" || artist == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	err := playlistManager.RemoveSongFromUserPlaylist(userID, playlistID, title, artist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("[Info] User %s removed song from playlist %s: %s - %s\n", userID, playlistID, artist, title)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "移除成功",
	})
}

// HandleDeleteUserPlaylist handles deleting a playlist
func HandleDeleteUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistID := r.URL.Query().Get("playlist_id")
	if playlistID == "" {
		http.Error(w, "Missing playlist_id parameter", http.StatusBadRequest)
		return
	}

	err := playlistManager.DeleteUserPlaylist(userID, playlistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("[Info] User %s deleted playlist %s\n", userID, playlistID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "删除成功",
	})
}

