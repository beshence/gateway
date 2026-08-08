package memory

import (
	"time"
)

type Bank struct {
	RootPublicKeyBytes []byte
	LeafPublicKeyBytes []byte
	LeafSignatureBytes []byte
}

type Challenge struct {
	Nonce     []byte
	ExpiresAt time.Time
}

type BankApiUrls struct {
	Urls []string
}
