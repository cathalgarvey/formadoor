//Package doorapi implements a client for github.com/formalabs/formadoor/clitools/doorMicroservice
package doorapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// Door represents a Door API endpoint and the shared secret needed to make
// authenticated calls.
type Door struct {
	Hostname string
	Port     int
	Secret   string
}

// DefaultDoor assumes you're running the API on the same machine on port 8080.
func DefaultDoor(secret string) *Door {
	return &Door{Hostname: "localhost", Port: 8080, Secret: secret}
}

func (door Door) addr() string {
	host := door.Hostname
	if host == "" {
		host = "localhost"
	}
	return "http://" + host + ":" + strconv.Itoa(door.Port) + "/"
}

// InstructDoorToOpenForSeconds - This call blocks because it's expected to be
// calling a localhost server, so any lag waiting for server response should
// be minimal.
func (door Door) InstructDoorToOpenForSeconds(seconds int) error {
	r, err := makeAPIRequestJSON(seconds)
	if err != nil {
		return err
	}
	bsecret, err := base64.StdEncoding.DecodeString(door.Secret)
	if err != nil {
		return err
	}
	mac := MakeB64MAC(r, bsecret)
	req, err := http.NewRequest("POST", door.addr(), bytes.NewReader(r))
	if err != nil {
		return err
	}
	req.Header.Set("hmac", mac)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}

type jsonAPIRequest struct {
	Seconds int
	When    int64 // Unix time, must be recent for validity
}

func makeAPIRequestJSON(seconds int) ([]byte, error) {
	r := jsonAPIRequest{
		Seconds: seconds,
		When:    time.Now().UTC().Unix(),
	}
	return json.Marshal(r)
}

// MakeB64MAC returns a b64-stringified MAC for message and key.
func MakeB64MAC(message, key []byte) (messageMAC string) {
	return base64.StdEncoding.EncodeToString(MakeMAC(message, key))
}

// MakeMAC returns a MAC for the given message and key.
func MakeMAC(message, key []byte) (messageMAC []byte) {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return expectedMAC
}
