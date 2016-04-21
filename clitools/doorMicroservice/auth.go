package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

var (
	// ErrNoMAC is returned if no MAC header is present
	ErrNoMAC = errors.New("No mac header found!")

	// ErrNoMatchingMACKey returned if no authKey can authenticate a body
	ErrNoMatchingMACKey = errors.New("No key can authenticate this request")

	// ErrOutdatedMAC is returned if a MAC validates but the request is outdated
	ErrOutdatedMAC = errors.New("MAC is valid but request timestamp is out of date")
)

type jsonAPIRequest struct {
	Seconds int
	When    int64 // Unix time, must be recent for validity
}

func (jar jsonAPIRequest) validTime() bool {
	return (time.Now().UTC().Unix() < jar.When+5)
}

func getAuthenticatedBody(r *http.Request) (*jsonAPIRequest, error) {
	var jar jsonAPIRequest
	bodyContents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	bodyMacB64 := r.Header.Get("hmac")
	if bodyMacB64 == "" {
		return nil, ErrNoMAC
	}
	bodyMac, err := base64.StdEncoding.DecodeString(bodyMacB64)
	if err != nil {
		return nil, err
	}
	var authedKey authApp
	for _, appKey := range authKeys {
		thisKey, kerr := base64.StdEncoding.DecodeString(appKey.Key)
		if kerr != nil {
			return nil, kerr
		}
		if CheckMAC(bodyContents, bodyMac, thisKey) {
			authedKey = appKey
			goto MACAuthed
		}
	}
	return nil, ErrNoMatchingMACKey
MACAuthed:
	log15.Info("Key authorised by MAC", log15.Ctx{"key": authedKey})
	err = json.Unmarshal(bodyContents, &jar)
	if err != nil {
		return nil, err
	}
	if !jar.validTime() {
		return nil, ErrOutdatedMAC
	}
	log15.Info("Timestamp accepted for authenticated message", log15.Ctx{"key": authedKey, "jsonRequest": jar})
	return &jar, nil
}

// CheckMAC reports whether messageMAC is a valid HMAC tag for message.
func CheckMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
