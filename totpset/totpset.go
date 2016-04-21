package totpset

import (
	"errors"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
)

var (
	// ErrRateLimited is returned when an attempt is dropped because a Set is
	// in rate limiting mode.
	ErrRateLimited = errors.New("Rate limiting, authentication attempt ignored")

	// ErrInvalidCode is what it sounds like
	ErrInvalidCode = errors.New("Invalid code, rate limiting")
)

type Key struct {
	Secret string
	Name   string
	// Bag for stuff like email address, name, phone number, other such details.
	// Implementing code can set and retrieve data from here.
	Metadata map[string]interface{}
}

func NewKey(Name, Secret string) *Key {
	k := new(Key)
	k.Name = Name
	k.Secret = Secret
	k.Metadata = make(map[string]interface{})
	return k
}

func (k *Key) ValidateWithWGAndCallback(passcode string, wg *sync.WaitGroup, callback func(bool, *Key)) {
	defer wg.Done()
	if totp.Validate(passcode, k.Secret) {
		callback(true, k)
	}
}

type Set struct {
	Keys              []*Key
	RateLimitDuration time.Duration
	NoAttemptsUntil   time.Time
	ValidityCallback  func(validated *Key, passcode string) (ok bool, reason string)
}

// NewSet returns a prepared Set with the given seconds of rate limiting.
func NewSet(rateLimitDurationSeconds int, keys ...*Key) *Set {
	return &Set{
		Keys:              keys,
		ValidityCallback:  nil,
		RateLimitDuration: time.Second * time.Duration(rateLimitDurationSeconds),
		NoAttemptsUntil:   time.Now().Add(time.Second * -1),
	}
}

// Validate returns either a validated key and no error (great!),
// or a key and an error preventing validation, or nil if validation simply
// fails.
// To avoid ambiguity, it also returns a boolean representing the final result;
// validated or no.
// If validation fails, then NoAttemptsUntil is set until <RateLimitDuration>
// from now.
func (set *Set) Validate(passcode string, logCallback func(string)) (bool, *Key, error) {
	if logCallback == nil {
		logCallback = func(s string) {}
	}
	logCallback("Testing passcode " + passcode + " against key set.")
	if time.Now().Before(set.NoAttemptsUntil) {
		logCallback("Rate limited, validation aborted.")
		return false, nil, ErrRateLimited
	}
	wg := new(sync.WaitGroup)
	wg.Add(1) // So it doesn't return on the below wait immediately; this is
	// .Done()'d after the below for-loop.
	// Resulting Key is sent on this channel, or it's closed.
	c := make(chan *Key)
	// When all keys have had a chance to validate, if none succeed the channel
	// is closed and will therefore return nil.
	go func(wg *sync.WaitGroup, c chan *Key) {
		wg.Wait()
		close(c)
	}(wg, c)
	// Callback sent to keys when validating. It ensures the channel only sends
	// once, but logs if duplicate keys validate for the provided code (important
	// for accurate access logging)
	someone_validated := false
	cb := func(validated bool, key *Key) {
		if validated || (!someone_validated) {
			logCallback("Key validated: " + key.Name)
			c <- key
			someone_validated = true
		} else {
			logCallback("Additional key validated: " + key.Name)
		}
	}
	// Dispatch validation challenge an callback to keys.
	for _, k := range set.Keys {
		wg.Add(1)
		go k.ValidateWithWGAndCallback(passcode, wg, cb)
	}
	wg.Done() // To decrement by one and await the goroutines.
	// Receive either a key or nil when the waitgroup returns and c is closed.
	result := <-c
	if result == nil {
		logCallback("No matching valid code found for: " + passcode)
		set.RateLimit()
		return false, nil, ErrInvalidCode
	} else {
		// Key success!
		// If a callback, do that to be sure.
		if set.ValidityCallback != nil {
			if ok, reason := set.ValidityCallback(result, passcode); !ok {
				logCallback("Validated for '" + result.Name + "' but not authorised: " + reason)
				set.RateLimit()
				return false, result, errors.New(reason)
			}
		}
		// No callback; we're good to go.
		logCallback("Authenticated: " + result.Name)
		return true, result, nil
	}
}

// RateLimit sets this TOTPSet to reject input for the next few seconds (as configured)
func (set *Set) RateLimit() {
	set.NoAttemptsUntil = time.Now().Add(set.RateLimitDuration)
}
