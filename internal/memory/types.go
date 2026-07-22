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

type ApiUrls struct {
	ApiUrls   []string  `json:"apiUrls"`
	UpdatedAt time.Time `json:"updatedAt"`
}
