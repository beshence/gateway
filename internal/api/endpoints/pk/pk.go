package pk

import (
	"encoding/base64"
	"gateway/internal/memory"
	"gateway/internal/misc"
	"net/http"

	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
	"github.com/cloudflare/circl/sign/slhdsa"
	"github.com/gin-gonic/gin"
)

func GetPublicKeysV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		bank, ok := memory.Banks[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err":    "NO_BANK",
				"errmsg": "we don't have public key of this bank",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"err": "0",
			"root": gin.H{
				"pk": base64.RawURLEncoding.EncodeToString(bank.RootPublicKeyBytes),
			},
			"leaf": gin.H{
				"pk":  base64.RawURLEncoding.EncodeToString(bank.LeafPublicKeyBytes),
				"sig": base64.RawURLEncoding.EncodeToString(bank.LeafSignatureBytes),
			},
		})
	}
}

type postPublicKeysV1Request struct {
	Root postPublicKeysRootV1Request `json:"root" binding:"required"`
	Leaf postPublicKeysLeafV1Request `json:"leaf" binding:"required"`
}

type postPublicKeysRootV1Request struct {
	PublicKeyB64 string `json:"pk" binding:"required"`
}

type postPublicKeysLeafV1Request struct {
	PublicKeyB64 string `json:"pk" binding:"required"`
	SignatureB64 string `json:"sig" binding:"required"`
}

func PostPublicKeysV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		var req postPublicKeysV1Request

		if c.BindJSON(&req) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_BIND_JSON",
				"errmsg": "could not bind request body",
			})
			return
		}

		// decode root public key

		rootPublicKeyBytes, err := base64.RawURLEncoding.DecodeString(req.Root.PublicKeyB64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_ROOT_PK",
				"errmsg": "could not decode root public key",
			})
			return
		}

		rootPublicKey := slhdsa.PublicKey{ID: slhdsa.SHAKE_256s}

		err = rootPublicKey.UnmarshalBinary(rootPublicKeyBytes)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_ROOT_PK",
				"errmsg": "could not decode root public key",
			})
			return
		}

		// check bank id

		generatedBankID, _ := misc.GetBankID(rootPublicKey)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_ROOT_PK",
				"errmsg": "could not decode root public key",
			})
			return
		}

		if bankID != generatedBankID {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_ROOT_PK",
				"errmsg": "could not decode root public key",
			})
			return
		}

		// check leaf public key

		leafPublicKeyBytes, err := base64.RawURLEncoding.DecodeString(
			req.Leaf.PublicKeyB64,
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": "CANT_DECODE_LEAF_PK",
			})
			return
		}

		leafPublicKey := mldsa87.PublicKey{}

		if err := leafPublicKey.UnmarshalBinary(leafPublicKeyBytes); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": "INVALID_LEAF_PK",
			})
			return
		}

		// decode signature

		signature, err := base64.RawURLEncoding.DecodeString(
			req.Leaf.SignatureB64,
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": "CANT_DECODE_SIGNATURE",
			})
			return
		}

		// check signature

		domain := "BESHENCE-BANK-MLDSA-KEY-V1"

		message := make([]byte, 0, len(domain)+len(leafPublicKeyBytes))

		message = append(
			message,
			[]byte(domain)...,
		)

		message = append(
			message,
			leafPublicKeyBytes...,
		)

		valid := slhdsa.Verify(
			&rootPublicKey,
			slhdsa.NewMessage(message),
			signature,
			nil,
		)

		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": "INVALID_LEAF_SIGNATURE",
			})
			return
		}

		// store public keys

		memory.Mutex.Lock()

		memory.Banks[bankID] = memory.Bank{
			RootPublicKeyBytes: rootPublicKeyBytes,
			LeafPublicKeyBytes: leafPublicKeyBytes,
			LeafSignatureBytes: signature,
		}

		memory.Mutex.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"err": "0",
		})
	}
}
