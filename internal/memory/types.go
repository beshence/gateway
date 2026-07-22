package memory

import (
	"time"
)

type Bank struct {
	EK []byte
}

type Challenge struct {
	Ciphertext []byte
	Secret     []byte
	ExpiresAt  time.Time
}

type APIURLs struct {
	ApiUrls   []string  `json:"api_urls"`
	UpdatedAt time.Time `json:"updated_at"`
}
