package memory

import (
	"time"
)

type Bank struct {
	PublicKeyBytes []byte
}

type Challenge struct {
	Ciphertext []byte
	Secret     []byte
	ExpiresAt  time.Time
}

type BankApiUrls struct {
	Urls         []string
	PayloadB64   string
	SignatureB64 string
}
