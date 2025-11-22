package service

// Config represents the full configuration structure
type Config struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		Tls  struct {
			Enabled bool   `json:"enabled"`
			Cert    string `json:"cert_file"`
			Key     string `json:"key_file"`
		} `json:"tls"`
		Advanced struct {
			Readtimeout    int `json:"read_timeout"`
			Writetimeout   int `json:"write_timeout"`
			Idletimeout    int `json:"idle_timeout"`
			Maxheaderbytes int `json:"max_header_bytes"`
		} `json:"advanced"`
	} `json:"server"`
	Database struct {
		Driver   string `json:"driver"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Prefix   string `json:"prefix"`
	} `json:"database"`
}
