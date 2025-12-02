package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserStore manages user data
type UserStore struct {
	mu       sync.RWMutex
	Users    map[string]*User // key: user ID
	Usernames map[string]string // key: username, value: user ID
	Emails   map[string]string // key: email, value: user ID
	Sessions map[string]string // key: session token, value: user ID
	filePath string
}

var userStore *UserStore

// InitUserStore initializes the user store
func InitUserStore() {
	userStore = &UserStore{
		Users:     make(map[string]*User),
		Usernames: make(map[string]string),
		Emails:    make(map[string]string),
		Sessions:  make(map[string]string),
		filePath:  "./files/users.json",
	}

	// Create files directory if it doesn't exist
	os.MkdirAll("./files", 0755)

	// Load existing users
	userStore.loadFromFile()
}

// loadFromFile loads users from JSON file
func (us *UserStore) loadFromFile() error {
	us.mu.Lock()
	defer us.mu.Unlock()

	data, err := os.ReadFile(us.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	var storageUsers []*userStorage
	if err := json.Unmarshal(data, &storageUsers); err != nil {
		return err
	}

	// Convert from storage format to User struct and rebuild indexes
	for _, su := range storageUsers {
		user := &User{
			ID:        su.ID,
			Username:  su.Username,
			Email:     su.Email,
			Password:  su.Password, // Restore password
			CreatedAt: su.CreatedAt,
			UpdatedAt: su.UpdatedAt,
		}
		us.Users[user.ID] = user
		us.Usernames[strings.ToLower(user.Username)] = user.ID
		us.Emails[strings.ToLower(user.Email)] = user.ID
	}

	return nil
}

// userStorage is used for saving/loading users with passwords
type userStorage struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` // Include password for storage
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// saveToFile saves users to JSON file
// NOTE: Caller must hold the lock!
func (us *UserStore) saveToFile() error {
	// Convert map to slice of storage struct (with passwords)
	users := make([]*userStorage, 0, len(us.Users))
	for _, user := range us.Users {
		users = append(users, &userStorage{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(us.filePath, data, 0644)
}

// generateID generates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateToken generates a session token
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateUser creates a new user
func (us *UserStore) CreateUser(username, email, password string) (*User, error) {
	// Hash password BEFORE acquiring lock (this is slow!)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	// Check if username already exists
	if _, exists := us.Usernames[strings.ToLower(username)]; exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if _, exists := us.Emails[strings.ToLower(email)]; exists {
		return nil, fmt.Errorf("email already exists")
	}

	// Create user
	user := &User{
		ID:        generateID(),
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store user
	us.Users[user.ID] = user
	us.Usernames[strings.ToLower(username)] = user.ID
	us.Emails[strings.ToLower(email)] = user.ID

	// Save to file
	if err := us.saveToFile(); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user by username and password
func (us *UserStore) AuthenticateUser(username, password string) (*User, string, error) {
	us.mu.RLock()
	userID, exists := us.Usernames[strings.ToLower(username)]
	us.mu.RUnlock()

	if !exists {
		return nil, "", fmt.Errorf("invalid username or password")
	}

	us.mu.RLock()
	user := us.Users[userID]
	us.mu.RUnlock()

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid username or password")
	}

	// Generate session token
	token := generateToken()
	us.mu.Lock()
	us.Sessions[token] = user.ID
	us.mu.Unlock()

	return user, token, nil
}

// GetUserByToken retrieves a user by session token
func (us *UserStore) GetUserByToken(token string) (*User, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	userID, exists := us.Sessions[token]
	if !exists {
		return nil, fmt.Errorf("invalid session token")
	}

	user, exists := us.Users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// Logout removes a session token
func (us *UserStore) Logout(token string) {
	us.mu.Lock()
	defer us.mu.Unlock()
	delete(us.Sessions, token)
}

// HTTP Handlers

// HandleRegister handles user registration
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[API] Register request received from %s\n", r.RemoteAddr)
	
	if r.Method != http.MethodPost {
		fmt.Printf("[API] Register: Method not allowed: %s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Printf("[API] Reading request body...\n")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[API] Error reading body: %v\n", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	fmt.Printf("[API] Body received: %s\n", string(body))

	var req RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Printf("[API] Error parsing JSON: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("[API] Parsed: username=%s, email=%s\n", req.Username, req.Email)

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		fmt.Printf("[API] Validation failed: missing fields\n")
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		fmt.Printf("[API] Validation failed: password too short\n")
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Create user
	fmt.Printf("[API] Creating user...\n")
	user, err := userStore.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		fmt.Printf("[API] Error creating user: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("[API] User created successfully! ID=%s\n", user.ID)

	// Generate session token
	token := generateToken()
	userStore.mu.Lock()
	userStore.Sessions[token] = user.ID
	userStore.mu.Unlock()

	// Initialize user's favorite playlist
	if playlistManager != nil {
		playlistManager.InitializeUserPlaylists(user.ID)
	}

	fmt.Printf("[Info] User registered: %s\n", user.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  *user,
	})
}

// HandleLogin handles user login
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[API] Login request received from %s\n", r.RemoteAddr)
	
	if r.Method != http.MethodPost {
		fmt.Printf("[API] Login: Method not allowed: %s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Printf("[API] Reading login body...\n")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[API] Error reading body: %v\n", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	fmt.Printf("[API] Login body: %s\n", string(body))

	var req LoginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Printf("[API] Error parsing JSON: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("[API] Login username: %s\n", req.Username)

	// Authenticate user
	fmt.Printf("[API] Authenticating...\n")
	user, token, err := userStore.AuthenticateUser(req.Username, req.Password)
	fmt.Printf("[API] Authentication result: %v\n", err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Printf("[Info] User logged in: %s\n", user.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  *user,
	})
}

// HandleLogout handles user logout
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	userStore.Logout(token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Logged out successfully",
	})
}

// HandleGetCurrentUser returns current user info
func HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	user, err := userStore.GetUserByToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// AuthMiddleware authenticates requests
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 尝试从 Authorization Header 获取
		token := r.Header.Get("Authorization")
		if token != "" {
			token = strings.TrimPrefix(token, "Bearer ")
			user, err := userStore.GetUserByToken(token)
			if err == nil {
				r.Header.Set("X-User-ID", user.ID)
				next(w, r)
				return
			}
		}

		// 2. 尝试从 X-Device-Token 获取 (ESP32设备)
		deviceToken := r.Header.Get("X-Device-Token")
		if deviceToken != "" {
			// 这里需要调用DeviceManager，但为了避免循环依赖（如果user包独立），我们使用全局函数或接口
			// 假设GetDeviceManager()是全局可用的
			dm := GetDeviceManager()
			device, err := dm.VerifyToken(deviceToken)
			if err == nil {
				// 找到了设备，现在需要找到对应的用户
				// Device结构体中有Username，我们需要根据Username找到User对象
				userID := userStore.GetUserIDByUsername(device.Username)
				if userID != "" {
					r.Header.Set("X-User-ID", userID)
					
					// 更新设备最后在线时间
					dm.UpdateLastSeen(device.MAC)
					
					next(w, r)
					return
				}
			}
		}

		// 3. 尝试从 Cookie 获取 (Web Session)
		cookie, err := r.Cookie("session_token")
		if err == nil {
			username := userStore.GetUsernameByToken(cookie.Value)
			if username != "" {
				userID := userStore.GetUserIDByUsername(username)
				if userID != "" {
					r.Header.Set("X-User-ID", userID)
					next(w, r)
					return
				}
			}
		}

		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
	}
}

// GetUserStore 返回UserStore单例
func GetUserStore() *UserStore {
	return userStore
}

// GetUserIDByUsername returns user ID by username
func (us *UserStore) GetUserIDByUsername(username string) string {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return us.Usernames[username]
}

// GetUsernameByToken 通过session token获取用户名
func (us *UserStore) GetUsernameByToken(token string) string {
	us.mu.RLock()
	defer us.mu.RUnlock()

	userID, exists := us.Sessions[token]
	if !exists {
		return ""
	}

	user, exists := us.Users[userID]
	if !exists {
		return ""
	}

	return user.Username
}
