package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ListFiles function: Traverse all files in the specified directory and return a slice of the file path
func ListFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Get Content function: Read the content of a specified file and return it
func GetFileContent(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get File Size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()

	// Read File Content
	fileContent := make([]byte, fileSize)
	_, err = file.Read(fileContent)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}

// fileHandler function: Handle file requests
func fileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "MeowMusicEmbeddedServer")
	filePath := r.URL.Path

	// 提前URL解码
	decodedPath, decodeErr := url.QueryUnescape(filePath)
	if decodeErr == nil {
		// 兼容历史数据：将+替换为空格（仅当解码成功时）
		decodedPath = strings.ReplaceAll(decodedPath, "+", " ")
		filePath = decodedPath // 后续统一使用解码后路径
	}

	// 处理 /url/ 远程请求（保持不变）
	if strings.HasPrefix(filePath, "/url/") {
		// Extract the URL after "/url/"
		urlPath := filePath[len("/url/"):]
		// Decode the URL path in case it's URL encoded
		decodedURL, err := url.QueryUnescape(urlPath)
		if err != nil {
			NotFoundHandler(w, r)
			return
		}
		// Determine the protocol based on the URL path
		var protocol string
		if strings.HasPrefix(decodedURL, "http/") {
			protocol = "http://"
		} else if strings.HasPrefix(decodedURL, "https/") {
			protocol = "https://"
		} else {
			NotFoundHandler(w, r)
			return
		}
		// Remove the protocol part from the decoded URL
		decodedURL = strings.TrimPrefix(decodedURL, "http/")
		decodedURL = strings.TrimPrefix(decodedURL, "https/")
		// Correctly concatenate the protocol with the decoded URL
		decodedURL = protocol + decodedURL
		// Create a new HTTP request to the decoded URL, without copying headers
		req, err := http.NewRequest("GET", decodedURL, nil)
		if err != nil {
			NotFoundHandler(w, r)
			return
		}
		// Send the request and get the response
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			NotFoundHandler(w, r)
			return
		}
		defer resp.Body.Close()
		// Read the response body into a byte slice
		fileContent, err := io.ReadAll(resp.Body)
		if err != nil {
			NotFoundHandler(w, r)
			return
		}
		setContentType(w, decodedURL)
		// Write file content to response
		w.Write(fileContent)
		return
	}

	// 统一使用解码后路径
	fullPath := filepath.Join("./files", filePath)
	fileContent, err := GetFileContent(fullPath)

	// 特殊处理空music.mp3
	isEmptyMusic := (err == nil && len(fileContent) == 0 && strings.HasSuffix(filePath, "/music.mp3"))
	if err != nil || isEmptyMusic {
		// 没有/空的music.mp3文件，直接返回404
		NotFoundHandler(w, r)
		return
	}

	// 避免重复Content-Type设置
	setContentType(w, filePath)
	w.Write(fileContent)
}

func setContentType(w http.ResponseWriter, path string) {
	ext := strings.ToLower(filepath.Ext(path))
	contentTypes := map[string]string{
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		".amr":  "audio/amr",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".txt":  "text/plain",
		".lrc":  "text/plain",
		".mrc":  "text/plain",
		".json": "application/json",
	}
	if ct, ok := contentTypes[ext]; ok {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}
