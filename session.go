package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type CustomSession struct {
	ID         string
	ExpiryTime int64
	Provider   string
	UserData   string
}

func NewCustomSession() CustomSession {
	res := CustomSession{}
	rb, _ := generateRandomBytes(32)
	res.ID = string(rb)
	res.ExpiryTime = time.Now().Add(6 * time.Hour).Unix()
	res.Provider = ""
	res.UserData = ""
	return res
}

var _sessionSvrToken = ""

func GetSessionSvrToken() string {
	if _sessionSvrToken == "" {
		_sessionSvrToken = os.Getenv("SESSION_SERVER_TOKEN")
	}
	return _sessionSvrToken
}

var _sessionCookieName = ""

func GetSessionCookieName() string {
	if _sessionCookieName == "" {
		_sessionCookieName = fmt.Sprintf("_session%s", os.Getenv("SESSION_COOKIE_NAME"))
	}
	return _sessionCookieName
}

func createHash(key string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func Encrypt(data string, passphrase string) (string, error) {
	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return b64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(data string, passphrase string) ([]byte, error) {
	sDec, err := b64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := sDec[:nonceSize], sDec[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func CheckCookie(request *http.Request) bool {
	bRes := false

	debug("CheckCookie: Starting...")

	cookie, err := request.Cookie(GetSessionCookieName())
	if err != nil {
		debug("CheckCookie: %#v", err)
		return false
	}

	if len(cookie.Value) > 0 {
		decString, err := Decrypt(cookie.Value, GetSessionSvrToken())
		if err != nil {
			return false
		}

		var sess CustomSession
		err = json.Unmarshal(decString, &sess)
		if err != nil {
			return false
		}

		if sess.ExpiryTime > time.Now().Unix() {
			bRes = true
		}
	}

	return bRes
}

func AddCookie(request *http.Request, response *http.Response, provider string, userdata string) {
	debug("AddCookie: Starting...\n")

	if CheckCookie(request) == false {
		sess := NewCustomSession()
		sess.Provider = provider
		sess.UserData = userdata
		b, err := json.Marshal(sess)
		if err != nil {
			return
		}

		encString, err := Encrypt(string(b), GetSessionSvrToken())
		if err != nil {
			return
		}

		expiryTime := time.Now().Add(6 * time.Hour)
		cookie := &http.Cookie{Name: GetSessionCookieName(), Value: encString, Expires: expiryTime, Path: "/"}
		response.Header.Add("Set-Cookie", cookie.String())

		debug("AddCookie: Setting '%s'\n", GetSessionCookieName())
	}
}
