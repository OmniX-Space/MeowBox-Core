package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Device 设备信息结构
type Device struct {
	MAC         string    `json:"mac"`
	Username    string    `json:"username"`
	DeviceName  string    `json:"device_name"`
	Token       string    `json:"token"`
	BindTime    time.Time `json:"bind_time"`
	LastSeen    time.Time `json:"last_seen"`
	IsActive    bool      `json:"is_active"`
}

// BindingCode 绑定码结构
type BindingCode struct {
	Code       string    `json:"code"`
	Username   string    `json:"username"`
	ExpiresAt  time.Time `json:"expires_at"`
	Used       bool      `json:"used"`
}

// DeviceManager 设备管理器
type DeviceManager struct {
	devices      map[string]*Device      // MAC -> Device
	bindingCodes map[string]*BindingCode // Code -> BindingCode
	tokens       map[string]string       // Token -> MAC
	mu           sync.RWMutex
	filePath     string
}

var deviceManager *DeviceManager
var deviceManagerOnce sync.Once

// GetDeviceManager 获取设备管理器单例
func GetDeviceManager() *DeviceManager {
	deviceManagerOnce.Do(func() {
		deviceManager = &DeviceManager{
			devices:      make(map[string]*Device),
			bindingCodes: make(map[string]*BindingCode),
			tokens:       make(map[string]string),
			filePath:     "./devices.json",
		}
		deviceManager.LoadFromFile()
	})
	return deviceManager
}

// GenerateBindingCode 生成6位数字绑定码
func (dm *DeviceManager) GenerateBindingCode(username string) (string, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 生成6位随机数字
	var code string
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}

	// 检查是否已存在（小概率）
	if _, exists := dm.bindingCodes[code]; exists {
		// 递归重新生成
		return dm.GenerateBindingCode(username)
	}

	// 创建绑定码，5分钟有效
	bindingCode := &BindingCode{
		Code:      code,
		Username:  username,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	dm.bindingCodes[code] = bindingCode
	fmt.Printf("[Device] Generated binding code %s for user %s\n", code, username)

	return code, nil
}

// generateDeviceToken 生成设备Token
func generateDeviceToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// BindDevice 绑定设备
func (dm *DeviceManager) BindDevice(mac, bindingCode, deviceName string) (*Device, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 验证绑定码
	code, exists := dm.bindingCodes[bindingCode]
	if !exists {
		return nil, fmt.Errorf("绑定码不存在")
	}

	if code.Used {
		return nil, fmt.Errorf("绑定码已使用")
	}

	if time.Now().After(code.ExpiresAt) {
		return nil, fmt.Errorf("绑定码已过期")
	}

	// 检查设备是否已绑定
	if existingDevice, exists := dm.devices[mac]; exists {
		return nil, fmt.Errorf("设备已绑定到用户 %s", existingDevice.Username)
	}

	// 生成Token
	token := generateDeviceToken()

	// 创建设备
	device := &Device{
		MAC:        mac,
		Username:   code.Username,
		DeviceName: deviceName,
		Token:      token,
		BindTime:   time.Now(),
		LastSeen:   time.Now(),
		IsActive:   true,
	}

	// 保存设备信息
	dm.devices[mac] = device
	dm.tokens[token] = mac

	// 标记绑定码已使用
	code.Used = true

	// 保存到文件
	dm.SaveToFile()

	fmt.Printf("[Device] Device %s bound to user %s\n", mac, code.Username)

	return device, nil
}

// VerifyToken 验证设备Token
func (dm *DeviceManager) VerifyToken(token string) (*Device, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	mac, exists := dm.tokens[token]
	if !exists {
		return nil, fmt.Errorf("无效的Token")
	}

	device, exists := dm.devices[mac]
	if !exists {
		return nil, fmt.Errorf("设备不存在")
	}

	if !device.IsActive {
		return nil, fmt.Errorf("设备已停用")
	}

	return device, nil
}

// GetDeviceByMAC 根据MAC地址获取设备
func (dm *DeviceManager) GetDeviceByMAC(mac string) (*Device, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	device, exists := dm.devices[mac]
	if !exists {
		return nil, fmt.Errorf("设备未绑定")
	}

	return device, nil
}

// UpdateLastSeen 更新设备最后在线时间
func (dm *DeviceManager) UpdateLastSeen(mac string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if device, exists := dm.devices[mac]; exists {
		device.LastSeen = time.Now()
		dm.SaveToFile()
	}
}

// DirectBindDevice 直接绑定设备（不需要绑定码）
func (dm *DeviceManager) DirectBindDevice(mac, username, deviceName string) (*Device, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查设备是否已绑定
	if existingDevice, exists := dm.devices[mac]; exists {
		return nil, fmt.Errorf("设备已绑定到用户 %s", existingDevice.Username)
	}

	// 生成Token
	token := generateDeviceToken()

	// 创建设备
	device := &Device{
		MAC:        mac,
		Username:   username,
		DeviceName: deviceName,
		Token:      token,
		BindTime:   time.Now(),
		LastSeen:   time.Now(),
		IsActive:   true,
	}

	dm.devices[mac] = device
	dm.tokens[token] = mac

	dm.SaveToFile()

	fmt.Printf("[Device] Device %s directly bound to user %s\n", mac, username)
	return device, nil
}

// UnbindDevice 解绑设备
func (dm *DeviceManager) UnbindDevice(mac string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	device, exists := dm.devices[mac]
	if !exists {
		return fmt.Errorf("设备不存在")
	}

	// 删除Token映射
	delete(dm.tokens, device.Token)
	// 删除设备
	delete(dm.devices, mac)

	dm.SaveToFile()

	fmt.Printf("[Device] Device %s unbound\n", mac)
	return nil
}

// GetDevice 获取单个设备信息
func (dm *DeviceManager) GetDevice(mac string) (*Device, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	device, exists := dm.devices[mac]
	if !exists {
		return nil, fmt.Errorf("设备不存在")
	}

	return device, nil
}

// GetUserDevices 获取用户的所有设备
func (dm *DeviceManager) GetUserDevices(username string) []*Device {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var devices []*Device
	for _, device := range dm.devices {
		if device.Username == username {
			devices = append(devices, device)
		}
	}

	return devices
}

// SaveToFile 保存设备信息到文件
func (dm *DeviceManager) SaveToFile() error {
	data := struct {
		Devices map[string]*Device `json:"devices"`
	}{
		Devices: dm.devices,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("[Error] Failed to marshal devices:", err)
		return err
	}

	err = os.WriteFile(dm.filePath, jsonData, 0644)
	if err != nil {
		fmt.Println("[Error] Failed to write devices.json:", err)
		return err
	}

	return nil
}

// LoadFromFile 从文件加载设备信息
func (dm *DeviceManager) LoadFromFile() error {
	file, err := os.Open(dm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[Info] devices.json not found, creating new file")
			return nil
		}
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var fileData struct {
		Devices map[string]*Device `json:"devices"`
	}

	err = json.Unmarshal(data, &fileData)
	if err != nil {
		return err
	}

	dm.devices = fileData.Devices
	if dm.devices == nil {
		dm.devices = make(map[string]*Device)
	}

	// 重建Token索引
	for mac, device := range dm.devices {
		dm.tokens[device.Token] = mac
	}

	fmt.Printf("[Info] Loaded %d devices from file\n", len(dm.devices))

	return nil
}

// HTTP处理器

// GenerateBindingCodeHandler 生成绑定码
func GenerateBindingCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取当前登录用户
	username := GetCurrentUser(r) // 需要从session获取
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dm := GetDeviceManager()
	code, err := dm.GenerateBindingCode(username)
	if err != nil {
		http.Error(w, "Failed to generate binding code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"code":    code,
		"expires_in": 300, // 5分钟
	})
}

// BindDeviceHandler ESP32设备绑定接口
func BindDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("[API] ESP32 device bind request received from", r.RemoteAddr)

	// 解析请求
	var req struct {
		MAC         string `json:"mac"`
		BindingCode string `json:"binding_code"`
		DeviceName  string `json:"device_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MAC == "" || req.BindingCode == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// 如果没有提供设备名称，使用默认名称
	if req.DeviceName == "" {
		req.DeviceName = "ESP32音乐播放器"
	}

	dm := GetDeviceManager()
	device, err := dm.BindDevice(req.MAC, req.BindingCode, req.DeviceName)
	if err != nil {
		fmt.Printf("[Error] Device bind failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "设备绑定成功",
		"token":   device.Token,
		"username": device.Username,
	})
}

// VerifyDeviceHandler 验证设备Token
func VerifyDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("X-Device-Token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}

	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	dm := GetDeviceManager()
	device, err := dm.VerifyToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 更新最后在线时间
	dm.UpdateLastSeen(device.MAC)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"device": map[string]interface{}{
			"mac":         device.MAC,
			"username":    device.Username,
			"device_name": device.DeviceName,
			"bind_time":   device.BindTime,
			"last_seen":   device.LastSeen,
		},
	})
}

// GetCurrentUser 从请求中获取当前登录用户
func GetCurrentUser(r *http.Request) string {
	userStore := GetUserStore()

	// 1. 尝试从 Authorization Header 获取 (Bearer Token)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		// 直接从UserStore验证Token
		user, err := userStore.GetUserByToken(token)
		if err == nil && user != nil {
			return user.Username
		}
	}

	// 2. 尝试从 Cookie 获取 session_token
	cookie, err := r.Cookie("session_token")
	if err == nil {
		username := userStore.GetUsernameByToken(cookie.Value)
		if username != "" {
			return username
		}
	}

	// 3. 尝试从 X-Device-Token Header 获取 (用于ESP32设备)
	deviceToken := r.Header.Get("X-Device-Token")
	if deviceToken != "" {
		dm := GetDeviceManager()
		device, err := dm.VerifyToken(deviceToken)
		if err == nil && device != nil {
			// 更新最后在线时间
			dm.UpdateLastSeen(device.MAC)
			return device.Username
		}
	}

	return ""
}

// DirectBindDeviceHandler Web端直接绑定设备（不需要绑定码）
func DirectBindDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取当前用户
	username := GetCurrentUser(r)
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未登录",
		})
		return
	}

	// 解析请求
	var req struct {
		MAC        string `json:"mac"`
		DeviceName string `json:"device_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	if req.MAC == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "MAC地址不能为空",
		})
		return
	}

	// 如果没有提供设备名称，使用默认名称
	if req.DeviceName == "" {
		req.DeviceName = "ESP32音乐播放器"
	}

	dm := GetDeviceManager()
	device, err := dm.DirectBindDevice(req.MAC, username, req.DeviceName)
	if err != nil {
		fmt.Printf("[Error] Direct bind failed: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "设备绑定成功",
		"device": map[string]interface{}{
			"mac":         device.MAC,
			"device_name": device.DeviceName,
			"bind_time":   device.BindTime,
		},
	})
}

// ListUserDevicesHandler 列出用户的所有设备
func ListUserDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取当前用户
	username := GetCurrentUser(r)
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未登录",
		})
		return
	}

	dm := GetDeviceManager()
	devices := dm.GetUserDevices(username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"devices": devices,
	})
}

// UnbindDeviceHandler 解绑设备
func UnbindDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取当前用户
	username := GetCurrentUser(r)
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未登录",
		})
		return
	}

	// 解析请求
	var req struct {
		MAC string `json:"mac"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	if req.MAC == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "MAC地址不能为空",
		})
		return
	}

	dm := GetDeviceManager()
	
	// 检查设备是否属于当前用户
	device, err := dm.GetDevice(req.MAC)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "设备不存在",
		})
		return
	}

	if device.Username != username {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无权解绑此设备",
		})
		return
	}

	// 解绑设备
	err = dm.UnbindDevice(req.MAC)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "设备已解绑",
	})
}

// SyncDeviceHandler ESP32用MAC地址同步Token（用于网页端绑定后自动获取token）
func SyncDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MAC string `json:"mac"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MAC == "" {
		http.Error(w, "Missing MAC address", http.StatusBadRequest)
		return
	}

	dm := GetDeviceManager()
	device, err := dm.GetDevice(req.MAC)
	
	if err != nil || device == nil {
		// 设备未绑定
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "设备未绑定",
		})
		return
	}

	// 更新最后在线时间
	dm.UpdateLastSeen(device.MAC)

	// 返回 token 和用户名
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"token":    device.Token,
		"username": device.Username,
		"message":  "同步成功",
	})
	
	fmt.Printf("[Info] Device %s synced token for user: %s\n", device.MAC, device.Username)
}
