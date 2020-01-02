package internal

import (
	"crypto/rand"
	"encoding/base64"
)

func generateRandomBytes(n int, asBase64 bool) ([]byte, error) {
	var (
		b   = make([]byte, n)
		err error
	)

	_, err = rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	if asBase64 {
		dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
		base64.StdEncoding.Encode(dst, b)

		return dst, err
	}

	return b, err
}
