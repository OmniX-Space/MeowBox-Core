# ESP32设备绑定功能使用指南

## 📋 功能概述

ESP32音乐播放器现在支持绑定到用户账号，实现个性化功能：
- ✅ 设备与用户账号绑定
- ✅ 安全的Token认证机制
- ✅ 通过语音命令完成绑定
- ✅ 为后续歌单功能奠定基础

---

## 🚀 快速开始

### **第1步：启动服务器**

```powershell
cd d:\esp32-music-server\Meow\MeowEmbeddedMusicServer
go run .
```

服务器将在 `http://localhost:2233` 启动

---

### **第2步：用户登录**

访问：`http://localhost:2233`

使用测试账号登录：
- 用户名：`test`
- 密码：`123456`

---

### **第3步：生成绑定码**

**方式1：使用命令行工具（临时）**

```powershell
# 向服务器请求生成绑定码
curl -X POST http://localhost:2233/api/device/generate-code `
  -H "Content-Type: application/json" `
  -H "Cookie: session_token=YOUR_SESSION_TOKEN"
```

**方式2：使用浏览器控制台（推荐）**

1. 登录后按 `F12` 打开开发者工具
2. 切换到 **Console** 标签
3. 输入以下代码：

```javascript
fetch('/api/device/generate-code', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'}
})
.then(r => r.json())
.then(data => {
  console.log('绑定码：', data.code);
  alert('绑定码：' + data.code + '\n有效期：5分钟');
});
```

4. 记下显示的6位数字绑定码，例如：`123456`

---

### **第4步：ESP32端绑定**

对ESP32说：

```
"小智，绑定设备，绑定码123456"
```

或者：

```
"小智，bind device, binding code is 123456"
```

ESP32会回复：

```
✅ 设备绑定成功！
已绑定到用户: test
```

---

### **第5步：验证绑定状态**

对ESP32说：

```
"小智，查询设备状态"
```

ESP32会显示：

```
📱 设备信息:

MAC地址: AA:BB:CC:DD:EE:FF
绑定状态: ✅ 已绑定
绑定用户: test
服务器验证: ✅ 通过
```

---

## 🔧 高级功能

### **解绑设备**

对ESP32说：

```
"小智，解绑设备"
```

### **自定义设备名称**

对ESP32说：

```
"小智，绑定设备，绑定码123456，设备名称客厅音响"
```

---

## 📊 数据存储

### **服务器端**

设备信息存储在：`d:\esp32-music-server\Meow\MeowEmbeddedMusicServer\devices.json`

```json
{
  "devices": {
    "AA:BB:CC:DD:EE:FF": {
      "mac": "AA:BB:CC:DD:EE:FF",
      "username": "test",
      "device_name": "ESP32音乐播放器",
      "token": "a1b2c3d4e5f6...",
      "bind_time": "2025-11-24T16:00:00Z",
      "last_seen": "2025-11-24T16:30:00Z",
      "is_active": true
    }
  }
}
```

### **ESP32端**

Token存储在NVS（非易失性存储）中：
- 命名空间：`device`
- Key: `token` - 设备Token
- Key: `username` - 绑定的用户名

---

## 🔌 API文档

### **1. 生成绑定码**

```
POST /api/device/generate-code
Authorization: 需要登录
```

**响应：**
```json
{
  "success": true,
  "code": "123456",
  "expires_in": 300
}
```

---

### **2. ESP32绑定设备**

```
POST /api/esp32/bind
Content-Type: application/json
```

**请求体：**
```json
{
  "mac": "AA:BB:CC:DD:EE:FF",
  "binding_code": "123456",
  "device_name": "ESP32音乐播放器"
}
```

**响应：**
```json
{
  "success": true,
  "message": "设备绑定成功",
  "token": "a1b2c3d4e5f6...",
  "username": "test"
}
```

---

### **3. 验证设备Token**

```
GET /api/esp32/verify
X-Device-Token: a1b2c3d4e5f6...
```

**响应：**
```json
{
  "success": true,
  "device": {
    "mac": "AA:BB:CC:DD:EE:FF",
    "username": "test",
    "device_name": "ESP32音乐播放器",
    "bind_time": "2025-11-24T16:00:00Z",
    "last_seen": "2025-11-24T16:30:00Z"
  }
}
```

---

## 🧪 测试流程

### **完整测试步骤**

1. **启动服务器**
   ```powershell
   go run .
   ```

2. **生成绑定码**（使用浏览器控制台）

3. **ESP32绑定**
   - 对ESP32说："小智，绑定设备，绑定码123456"
   - 观察ESP32日志：
     ```
     [DeviceManager] Starting device binding with code: 123456
     [DeviceManager] Sending bind request to: http://...
     [DeviceManager] Bind request status code: 200
     [DeviceManager] Device successfully bound to user: test
     ```

4. **查询状态**
   - 对ESP32说："小智，查询设备状态"

5. **验证服务器数据**
   - 检查 `devices.json` 文件是否包含设备信息

6. **测试解绑**
   - 对ESP32说："小智，解绑设备"
   - 再次查询状态，应显示未绑定

---

## 🐛 故障排除

### **问题1：绑定码无效**

**现象**：ESP32提示"绑定失败"

**解决方案**：
- 检查绑定码是否输入正确
- 绑定码有效期为5分钟，请重新生成
- 确认网络连接正常

---

### **问题2：设备已绑定**

**现象**：提示"设备已绑定到用户XXX"

**解决方案**：
- 先解绑设备：对ESP32说"小智，解绑设备"
- 或者在服务器端删除 `devices.json` 中的设备记录

---

### **问题3：网络连接失败**

**现象**：ESP32无法连接到服务器

**解决方案**：
- 检查ESP32是否连接到WiFi
- 确认服务器地址是否正确：`http://http-embedded-music.miao-lab.top:2233`
- 检查防火墙设置

---

## 📝 下一步

绑定功能完成后，可以继续开发：

✅ **阶段1：基础绑定** - 已完成
⏳ **阶段2：歌单系统** - 待开发
⏳ **阶段3：Web管理界面** - 待开发

---

## 💡 提示

- 绑定码每次生成后只能使用一次
- 一个设备只能绑定到一个用户
- Token存储在ESP32的NVS中，断电不丢失
- 可以通过解绑并重新绑定来更换用户

---

**🎉 享受您的个性化ESP32音乐播放器！**
