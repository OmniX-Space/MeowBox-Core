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
	Password struct {
		Memory      uint32 `json:"memory"`
		Iterations  uint32 `json:"iterations"`
		Parallelism uint8  `json:"parallelism"`
		SaltLength  uint32 `json:"salt_length"`
		KeyLength   uint32 `json:"key_length"`
	} `json:"password"`
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

// Argon2Params defines the parameters for Argon2id (adjustable based on server performance)
type Argon2Params struct {
	Memory      uint32 // Memory usage (KiB), recommended 64*1024 = 64MB
	Iterations  uint32 // Time cost, recommended 1-3
	Parallelism uint8  // Number of parallel threads, recommended 2-4
	SaltLength  uint32 // Salt length, recommended 16 bytes
	KeyLength   uint32 // Output hash length, recommended 32 bytes
}

// hashParts Used to parse stored hash strings
type HashParts struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltBase64  string
	HashBase64  string
}
