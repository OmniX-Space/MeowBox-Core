package main

import "time"

// MusicItem represents a music item.
type MusicItem struct {
	Title        string `json:"title"`
	Artist       string `json:"artist"`
	AudioURL     string `json:"audio_url"`
	AudioFullURL string `json:"audio_full_url"`
	M3U8URL      string `json:"m3u8_url"`
	LyricURL     string `json:"lyric_url"`
	CoverURL     string `json:"cover_url"`
	Duration     int    `json:"duration"`
	FromCache    bool   `json:"from_cache"`
	IP           string `json:"ip"`
}

// User represents a user account
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPlaylist represents a user's custom playlist
type UserPlaylist struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CoverURL    string      `json:"cover_url"`
	Songs       []MusicItem `json:"songs"`
	IsPublic    bool        `json:"is_public"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
