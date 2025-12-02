package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SearchResult represents a search result with frontend-compatible fields
type SearchResult struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	URL      string `json:"url"`       // For frontend compatibility
	Link     string `json:"link"`      // Alternative field
	LyricURL string `json:"lyric_url"` // Lyrics URL
	CoverURL string `json:"cover_url"`
	Duration int    `json:"duration"`  // Duration in seconds
}

// HandleSearch handles music search requests
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	query := r.URL.Query().Get("query")
	fmt.Printf("[API] Search request: query=%s\n", query)
	
	if query == "" {
		json.NewEncoder(w).Encode([]SearchResult{})
		return
	}
	
	// Search from multiple sources
	items := searchFromAllSources(query)
	
	// Convert to frontend-compatible format
	results := make([]SearchResult, len(items))
	for i, item := range items {
		// Use AudioFullURL as primary, fallback to AudioURL
		audioURL := item.AudioFullURL
		if audioURL == "" {
			audioURL = item.AudioURL
		}
		
		results[i] = SearchResult{
			Title:    item.Title,
			Artist:   item.Artist,
			URL:      audioURL,
			Link:     audioURL,
			LyricURL: item.LyricURL,
			CoverURL: item.CoverURL,
			Duration: item.Duration,
		}
		
		fmt.Printf("[API] Result %d: title=%s, artist=%s, url=%s, lyric=%s\n", i+1, item.Title, item.Artist, audioURL, item.LyricURL)
	}
	
	fmt.Printf("[API] Search completed: %d results found\n", len(results))
	json.NewEncoder(w).Encode(results)
}

// searchFromAllSources searches music from all available sources
func searchFromAllSources(query string) []MusicItem {
	var results []MusicItem
	
	// Source 1: Search from sources.json (local config)
	sources := readSources()
	for _, source := range sources {
		if containsIgnoreCase(source.Title, query) || containsIgnoreCase(source.Artist, query) {
			item := MusicItem{
				Title:        source.Title,
				Artist:       source.Artist,
				AudioURL:     source.AudioURL,
				AudioFullURL: source.AudioFullURL,
				M3U8URL:      source.M3U8URL,
				LyricURL:     source.LyricURL,
				CoverURL:     source.CoverURL,
				Duration:     source.Duration,
				FromCache:    false,
			}
			results = append(results, item)
			
			// Limit results to prevent too many items
			if len(results) >= 20 {
				break
			}
		}
	}
	
	// Source 2: Search from local files
	localItems := searchLocalMusic(query)
	results = append(results, localItems...)
	
	// If no local results, try API search
	if len(results) == 0 {
		apiResults := searchFromAPI(query)
		results = append(results, apiResults...)
	}
	
	// Limit total results
	if len(results) > 20 {
		results = results[:20]
	}
	
	return results
}

// searchLocalMusic searches music from local files
func searchLocalMusic(query string) []MusicItem {
	// This function would search from local music files
	// For now, return empty array
	// TODO: Implement local file search
	return []MusicItem{}
}

// searchFromAPI searches music from external APIs
func searchFromAPI(query string) []MusicItem {
	var results []MusicItem
	
	// Try different sources in priority order: 酷我 > 网易云 > 咪咕 > 百度
	// 酷我的搜索结果更准确，原唱版本更容易排在前面
	sources := []string{"kuwo", "netease", "migu", "baidu"}
	
	for _, source := range sources {
		item := YuafengAPIResponseHandler(source, query, "")
		if item.Title != "" {
			results = append(results, item)
			break // Return first successful result
		}
	}
	
	return results
}

// containsIgnoreCase checks if a string contains substring (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	var result []rune
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 32
		}
		result = append(result, r)
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(substr) == 0 || indexString(s, substr) >= 0
}

func indexString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
