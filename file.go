package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	// Obtain the path of the request
	filePath := r.URL.Path

	// Check if the request path starts with "/url/"
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
		// Set appropriate Content-Type based on file extension
		ext := filepath.Ext(decodedURL)
		switch ext {
		case ".mp3":
			w.Header().Set("Content-Type", "audio/mpeg")
		case ".wav":
			w.Header().Set("Content-Type", "audio/wav")
		case ".flac":
			w.Header().Set("Content-Type", "audio/flac")
		case ".aac":
			w.Header().Set("Content-Type", "audio/aac")
		case ".ogg":
			w.Header().Set("Content-Type", "audio/ogg")
		case ".m4a":
			w.Header().Set("Content-Type", "audio/mp4")
		case ".amr":
			w.Header().Set("Content-Type", "audio/amr")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".bmp":
			w.Header().Set("Content-Type", "image/bmp")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".webp":
			w.Header().Set("Content-Type", "image/webp")
		case ".txt":
			w.Header().Set("Content-Type", "text/plain")
		case ".lrc":
			w.Header().Set("Content-Type", "text/plain")
		case ".mrc":
			w.Header().Set("Content-Type", "text/plain")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		default:
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		// Write file content to response
		w.Write(fileContent)
		return
	}

	// Construct the complete file path
	fullFilePath := filepath.Join("./files", filePath)

	// Try replacing '+' with ' ' and check if the file exists
	tempFilePath := strings.ReplaceAll(fullFilePath, "+", " ")
	if _, err := os.Stat(tempFilePath); err == nil {
		fullFilePath = tempFilePath
	}

	// Get file content
	fileContent, err := GetFileContent(fullFilePath)
	
	// 检查是否为空文件（特别是 music.mp3，如果是空的也视为不存在，需要等待转码）
	if err == nil && len(fileContent) == 0 && strings.HasSuffix(filePath, "/music.mp3") {
		err = fmt.Errorf("file is empty")
	}
	
	if err != nil {
		// If file not found, try replacing ' ' with '+' and check again
		tempFilePath = strings.ReplaceAll(fullFilePath, " ", "+")
		fileContent, err = GetFileContent(tempFilePath)
		
		// 同样检查带 + 的路径是否为空
		if err == nil && len(fileContent) == 0 && strings.HasSuffix(filePath, "/music.mp3") {
			err = fmt.Errorf("file is empty")
		}

		if err != nil {
			// 特殊处理：如果请求的是缓存中的文件，等待后台处理完成
			fmt.Printf("[Web Access] File not found, checking path prefix: %s\n", filePath)
			if strings.HasPrefix(filePath, "/cache/music/") {
				// music.mp3 等待最多 60 秒，歌词等待最多 10 秒
				maxWait := 10
				if strings.HasSuffix(filePath, "/music.mp3") {
					maxWait = 60
				}
				
				// URL 解码路径（处理中文文件名）
				decodedPath, _ := url.QueryUnescape(filePath)
				decodedPath = strings.ReplaceAll(decodedPath, "+", " ") // + 转空格
				decodedFullPath := filepath.Join("./files", decodedPath)
				
				fmt.Printf("[Web Access] Waiting for file: %s (max %d seconds)\n", decodedFullPath, maxWait)
				for i := 0; i < maxWait; i++ {
					time.Sleep(1 * time.Second)
					// 检查 URL 解码后的路径
					if _, err := os.Stat(decodedFullPath); err == nil {
						fmt.Printf("[Web Access] File ready after %d seconds: %s\n", i+1, decodedFullPath)
						http.ServeFile(w, r, decodedFullPath)
						return
					}
					// 检查原始路径
					if _, err := os.Stat(fullFilePath); err == nil {
						http.ServeFile(w, r, fullFilePath)
						return
					}
					// 检查带 + 的路径
					tempPath := strings.ReplaceAll(fullFilePath, " ", "+")
					if _, err := os.Stat(tempPath); err == nil {
						http.ServeFile(w, r, tempPath)
						return
					}
				}
			}
			NotFoundHandler(w, r)
			return
		}
	}

	// Set appropriate Content-Type based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	case ".wav":
		w.Header().Set("Content-Type", "audio/wav")
	case ".flac":
		w.Header().Set("Content-Type", "audio/flac")
	case ".aac":
		w.Header().Set("Content-Type", "audio/aac")
	case ".ogg":
		w.Header().Set("Content-Type", "audio/ogg")
	case ".m4a":
		w.Header().Set("Content-Type", "audio/mp4")
	case ".amr":
		w.Header().Set("Content-Type", "audio/amr")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".bmp":
		w.Header().Set("Content-Type", "image/bmp")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".txt":
		w.Header().Set("Content-Type", "text/plain")
	case ".lrc":
		w.Header().Set("Content-Type", "text/plain")
	case ".mrc":
		w.Header().Set("Content-Type", "text/plain")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Write file content to response
	w.Write(fileContent)
}
