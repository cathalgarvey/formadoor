package totpset

import (
  "time"
  "testing"

  "github.com/stretchr/testify/assert"
  "github.com/pquerna/otp/totp"
)

var (
  secret1 string
  secret2 string
  testSet *Set
)

func init() {
  secretK1, _ := totp.Generate(totp.GenerateOpts{
    Issuer: "foo.bar",
    AccountName: "baz@foo.bar",
  })
  secret1 = secretK1.Secret()
  secretK2, _ := totp.Generate(totp.GenerateOpts{
    Issuer: "foo.bar",
    AccountName: "qux@foo.bar",
  })
  secret2 = secretK2.Secret()
  testSet = NewSet(5, NewKey("baz", secret1), NewKey("qux", secret2))
}

func TestTOTPValidation(t *testing.T) {
  var code string
  // Should definitely fail (wrong length)
  code = "11111"
  ok, match, err := testSet.Validate(code, nil)
  assert.False(t, ok)
  assert.Nil(t, match)
  assert.Equal(t, ErrInvalidCode, err)
  assert.True(t, testSet.NoAttemptsUntil.After(time.Now()))
  // Test rate limiting!
  ok, match, err = testSet.Validate(code, nil)
  assert.False(t, ok)
  assert.Nil(t, match)
  assert.Equal(t, ErrRateLimited, err)
  assert.True(t, testSet.NoAttemptsUntil.After(time.Now()))
  testSet.NoAttemptsUntil = time.Now().Add(time.Second * -1)
  // Should probably fail (not deliberately correct)
  code = "111111"
  ok, match, err = testSet.Validate(code, nil)
  assert.False(t, ok)
  assert.Nil(t, match)
  assert.Equal(t, ErrInvalidCode, err)
  assert.True(t, testSet.NoAttemptsUntil.After(time.Now()))
  testSet.NoAttemptsUntil = time.Now().Add(time.Second * -1)
  // Should succeed (generated from key, default skew should guarantee validity)
  code = totp.GenerateCode(secret1, time.Now())
  ok, match, err = testSet.Validate(code, nil)
  assert.Nil(t, err)
  assert.True(t, ok)
  assert.Equal(t, secret1, match.Secret)
  assert.False(t, testSet.NoAttemptsUntil.After(time.Now()))
}
