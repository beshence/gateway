package misc

import (
	"crypto/hmac"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha3"
	"encoding/base32"
	"encoding/base64"
	"hash"
	"strings"
)

func GetBankID(key *mlkem.EncapsulationKey1024) string {
	h := sha3.New256()

	h.Write([]byte("BESHENCE-BANK-ID-V1"))
	h.Write(key.Bytes())

	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	encodedStr := encoder.EncodeToString(h.Sum(nil))
	return strings.ToLower(encodedStr)
}

func RandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func MakeProof(key []byte, data []byte) []byte {
	h := hmac.New(
		func() hash.Hash { return sha3.New256() },
		key,
	)
	h.Write(data)
	return h.Sum(nil)
}
