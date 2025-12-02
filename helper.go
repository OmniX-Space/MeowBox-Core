package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Source is an alias for MusicItem (used in sources.json)
type Source = MusicItem

// Download file from URL
func downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Get IP address from request
func IPhandler(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip, nil
}

// Read sources from sources.json
func readSources() []Source {
	file, err := os.Open("sources.json")
	if err != nil {
		return []Source{}
	}
	defer file.Close()

	var sources []Source
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sources)
	if err != nil {
		return []Source{}
	}
	return sources
}

// Read music from cache
func readFromCache(path string) (MusicItem, bool) {
	// Logic to read music item from a cached folder path
	// This assumes path is like "files/cache/music/Artist-Song"
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return MusicItem{}, false
	}

	dirName := filepath.Base(path)
	parts := strings.SplitN(dirName, "-", 2)
	var artist, title string
	if len(parts) == 2 {
		artist = parts[0]
		title = parts[1]
	} else {
		title = dirName
	}

	return getLocalMusicItem(title, artist), true
}

// Request and cache music from API
func requestAndCacheMusic(song, singer string) MusicItem {
	// Try different sources in priority order
	sources := []string{"kuwo", "netease", "migu", "baidu"}
	for _, source := range sources {
		item := YuafengAPIResponseHandler(source, song, singer)
		if item.Title != "" {
			return item
		}
	}
	return MusicItem{}
}

// 直接从远程URL流式转码（边下载边转码，超快！）
func streamConvertAudio(inputURL, outputFile string) error {
	fmt.Printf("[Info] Stream converting from URL (fast mode)\n")

	// 先写入临时文件，完成后再重命名（避免读取到不完整的文件）
	tempFile := outputFile + ".tmp"

	// ffmpeg 直接读取远程 URL 并转码
	// -t 600: 只下载前10分钟，减少80%下载量！
	// 移除 reconnect 参数，避免兼容性问题
	// 添加 -bufsize 以提高稳定性
	cmd := exec.Command("ffmpeg", "-y",
		"-t", "600",
		"-i", inputURL,
		"-threads", "0",
		"-ac", "1", "-ar", "24000", "-b:a", "32k", "-q:a", "9",
		"-bufsize", "64k",
		tempFile)

	err := cmd.Run()
	if err != nil {
		fmt.Printf("[Error] Stream convert failed: %v\n", err)
		os.Remove(tempFile) // 清理临时文件
		return err
	}

	// 检查生成的文件大小
	fileInfo, err := os.Stat(tempFile)
	if err != nil || fileInfo.Size() < 1024 {
		fmt.Printf("[Error] Stream converted file is too small or empty\n")
		os.Remove(tempFile)
		return fmt.Errorf("converted file is too small")
	}

	// 转码完成后重命名为最终文件
	err = os.Rename(tempFile, outputFile)
	if err != nil {
		fmt.Printf("[Error] Failed to rename temp file: %v\n", err)
		return err
	}

	fmt.Printf("[Success] Stream convert completed: %s\n", outputFile)
	return nil
}

// 实时流式转码到 HTTP Writer（边下载边播放！）
func streamConvertToWriter(inputURL string, w http.ResponseWriter) error {
	fmt.Printf("[Info] Live streaming from URL: %s\n", inputURL)

	// ffmpeg 边下载边转码，输出到 stdout
	cmd := exec.Command("ffmpeg",
		"-i", inputURL,
		"-threads", "0",
		"-ac", "1", "-ar", "24000", "-b:a", "32k", "-q:a", "9",
		"-f", "mp3",
		"-map_metadata", "-1",
		"pipe:1") // 输出到 stdout

	// 获取 stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// 启动 ffmpeg
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	// 设置响应头
	w.Header().Set("Content-Type", "audio/mpeg")
	// 移除 Transfer-Encoding: chunked，让 Go 自动处理

	// 边读边写到 HTTP response
	buf := make([]byte, 8192)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			if f, ok := w.(http.Flusher); ok {
				f.Flush() // 立即发送给客户端
			}
		}
		if err != nil {
			break
		}
	}

	cmd.Wait()
	fmt.Printf("[Success] Live streaming completed\n")
	return nil
}

// Helper function for identifying file formats
func getMusicFileExtension(url string) (string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return "", err
	}
	// Get file format from Content-Type header
	contentType := resp.Header.Get("Content-Type")
	ext, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	// Identify file extension based on file format
	switch ext {
	case "audio/mpeg":
		return ".mp3", nil
	case "audio/flac":
		return ".flac", nil
	case "audio/x-flac":
		return ".flac", nil
	case "audio/wav":
		return ".wav", nil
	case "audio/aac":
		return ".aac", nil
	case "audio/ogg":
		return ".ogg", nil
	case "application/octet-stream":
		// Try to guess file format from URL or other information
		if strings.Contains(url, ".mp3") {
			return ".mp3", nil
		} else if strings.Contains(url, ".flac") {
			return ".flac", nil
		} else if strings.Contains(url, ".wav") {
			return ".wav", nil
		} else if strings.Contains(url, ".aac") {
			return ".aac", nil
		} else if strings.Contains(url, ".ogg") {
			return ".ogg", nil
		} else {
			return "", fmt.Errorf("unknown file format from Content-Type and URL: %s", contentType)
		}
	default:
		return "", fmt.Errorf("unknown file format: %s", ext)
	}
}

// Helper function for identifying file formats
func GetDuration(filePath string) int {
	fmt.Printf("[Info] Get duration of obtaining music file %s\n", filePath)
	// Use ffprobe to get audio duration
	output, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath).Output()
	if err != nil {
		fmt.Println("[Error] Error getting audio duration:", err)
		return 0
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		fmt.Println("[Error] Error converting duration to float:", err)
		return 0
	}

	return int(duration)
}

// Helper function to compress and segment audio file
func compressAndSegmentAudio(inputFile, outputDir string) error {
	fmt.Printf("[Info] Compress and segment audio file %s\n", inputFile)
	// Compress music files
	outputFile := filepath.Join(outputDir, "music.mp3")
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-ac", "1", "-ab", "32k", "-ar", "24000", outputFile)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Helper function to obtain music data from local folder
func getLocalMusicItem(song, singer string) MusicItem {
	musicDir := "./files/music"
	fmt.Println("[Info] Reading local folder music.")
	files, err := os.ReadDir(musicDir)
	if err != nil {
		fmt.Println("[Error] Failed to read local music directory:", err)
		return MusicItem{}
	}

	for _, file := range files {
		if file.IsDir() {
			if singer == "" {
				if strings.Contains(file.Name(), song) {
					dirPath := filepath.Join(musicDir, file.Name())
					// Extract artist and title from the directory name
					parts := strings.SplitN(file.Name(), "-", 2)
					var artist, title string
					if len(parts) == 2 {
						artist = parts[0]
						title = parts[1]
					} else {
						title = file.Name()
					}
					basePath := "/cache/music/" + url.QueryEscape(file.Name())
					return MusicItem{
						Title:        title,
						Artist:       artist,
						CoverURL:     basePath + "/cover.jpg",
						LyricURL:     basePath + "/lyric.lrc",
						AudioFullURL: basePath + "/music.mp3",
						AudioURL:     basePath + "/music.mp3",
						M3U8URL:      basePath + "/music.m3u8",
						Duration:     GetDuration(filepath.Join(dirPath, "music.mp3")),
					}
				}
			} else {
				if strings.Contains(file.Name(), song) && strings.Contains(file.Name(), singer) {
					dirPath := filepath.Join(musicDir, file.Name())
					// Extract artist and title from the directory name
					parts := strings.SplitN(file.Name(), "-", 2)
					var artist, title string
					if len(parts) == 2 {
						artist = parts[0]
						title = parts[1]
					} else {
						title = file.Name()
					}
					basePath := "/cache/music/" + url.QueryEscape(file.Name())
					return MusicItem{
						Title:        title,
						Artist:       artist,
						CoverURL:     basePath + "/cover.jpg",
						LyricURL:     basePath + "/lyric.lrc",
						AudioFullURL: basePath + "/music.mp3",
						AudioURL:     basePath + "/music.mp3",
						M3U8URL:      basePath + "/music.m3u8",
						Duration:     GetDuration(filepath.Join(dirPath, "music.mp3")),
					}
				}
			}
		}
	}
	return MusicItem{}
}
