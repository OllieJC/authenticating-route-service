package internal

import (
	c "authenticating-route-service/internal/configurator"
	u "authenticating-route-service/internal/utils"
	. "authenticating-route-service/pkg/debugprint"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	rb, _ := u.GenerateRandomBytes(16, true)
	res.ID = string(rb)
	res.ExpiryTime = expiryTime().Unix()
	res.Provider = ""
	res.UserData = ""
	return res
}

var _sessionSvrToken = ""

func GetSessionSvrToken(request *http.Request) string {
	if _sessionSvrToken == "" {
		dc, err := c.GetDomainConfigFromRequest(request)
		if err == nil && dc.Domain != "" {
			_sessionSvrToken = dc.SessionServerToken
		}
	}
	return _sessionSvrToken
}

var _sessionCookieName = ""

func GetSessionCookieName(request *http.Request) string {
	if _sessionCookieName == "" {
		dc, err := c.GetDomainConfigFromRequest(request)
		if err == nil && dc.Domain != "" {
			_sessionCookieName = fmt.Sprintf("_session%s", dc.SessionCookieName)
		}
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

func CheckCookie(request *http.Request) (bool, CustomSession) {

	var (
		blank_sess CustomSession
	)

	Debugfln("CheckCookie: Starting...")

	cookie, err := request.Cookie(GetSessionCookieName(request))
	if err != nil {
		Debugfln("CheckCookie: %#v", err)
		return false, blank_sess
	}

	if len(cookie.Value) > 0 {
		decString, err := Decrypt(cookie.Value, GetSessionSvrToken(request))
		if err != nil {
			Debugfln("CheckCookie: %#v", err)
			return false, blank_sess
		}

		var sess CustomSession
		err = json.Unmarshal(decString, &sess)
		if err != nil {
			Debugfln("CheckCookie: %#v", err)
			return false, blank_sess
		}

		Debugfln("CheckCookie: Session ID: %s Session Expiry: %d", sess.ID, sess.ExpiryTime)

		if sess.ExpiryTime > time.Now().Unix() {
			return true, sess
		}
	}

	Debugfln("CheckCookie: returning false")
	return false, blank_sess
}

func AddCookie(request *http.Request, response *http.Response, provider string, userdata string) {
	Debugfln("AddCookie: Starting...")

	ok, cookieSess := CheckCookie(request)

	sess := NewCustomSession()
	if ok {
		Debugfln("AddCookie: Session cookie already exists")
		sess.Provider = cookieSess.Provider
		sess.UserData = cookieSess.UserData
	} else {
		Debugfln("AddCookie: Cookie doesn't exist")
		sess.Provider = provider
		sess.UserData = userdata
	}

	Debugfln("AddCookie: Provider: %s, UserData length: %d", sess.Provider, len(sess.UserData))

	b, err := json.Marshal(sess)
	if err != nil {
		Debugfln("AddCookie: err: %#v", err)
		return
	}

	encString, err := Encrypt(string(b), GetSessionSvrToken(request))
	if err != nil {
		Debugfln("AddCookie: err: %#v", err)
		return
	}

	expiryTime := expiryTime()
	cookie := &http.Cookie{
		Name:     GetSessionCookieName(request),
		Value:    encString,
		Expires:  expiryTime,
		Path:     "/",
		HttpOnly: true,
	}
	response.Header.Add("Set-Cookie", cookie.String())

	Debugfln("AddCookie: Setting '%s'", GetSessionCookieName(request))

}

func expiryTime() time.Time {
	return time.Now().Add(6 * time.Hour)
}
