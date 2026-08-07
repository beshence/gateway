package memory

import (
	"time"
)

type Bank struct {
	PublicKeyBytes []byte
}

type Challenge struct {
	Nonce     []byte
	ExpiresAt time.Time
}

type BankApiUrls struct {
	Urls         []string
	PayloadB64   string
	SignatureB64 string
}
