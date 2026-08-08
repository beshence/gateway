package misc

import (
	"crypto/sha256"
	"encoding/base32"
	"strings"

	"github.com/cloudflare/circl/sign/slhdsa"
)

func GetBankID(publicKey slhdsa.PublicKey) (string, error) {
	publicKeyBytes, err := publicKey.MarshalBinary()

	if err != nil {
		return "", err
	}

	h := sha256.New()

	_, err = h.Write([]byte("BESHENCE-BANK-ID-V1"))

	if err != nil {
		return "", err
	}
	_, err = h.Write(publicKeyBytes)

	if err != nil {
		return "", err
	}

	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	encodedStr := encoder.EncodeToString(h.Sum(nil))

	return strings.ToLower(encodedStr), nil
}
