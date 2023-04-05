package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type crypto struct {
	key string
}

func New(key string) *crypto {
	return &crypto{key: key}
}

func (crypto *crypto) CheckHash(m metrics.Metric) bool {
	hash := crypto.CreateHash(m)
	received, _ := hex.DecodeString(m.Hash)
	if len(crypto.key) > 0 && !hmac.Equal(received, hash) {
		log.Println("crypto::checkHash::info: wrong hash")
		return false
	}
	return true
}

func (crypto *crypto) CreateHash(m metrics.Metric) []byte {
	h := hmac.New(sha256.New, []byte(crypto.key))
	if m.MType == "gauge" {
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)))
	} else {
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)))
	}
	return h.Sum(nil)
}
