package main

import (
	"time"

	"github.com/cathalgarvey/formadoor/totpset"
)

// Accepts a validated key and tests the associated time policy.
func passcodeToTimePolicy(validated *totpset.Key, passcode string) (ok bool, reason string) {
	// validated should have additional metadata "policy" (a Policy pointer) and "email" (a string)
	accountI, present := validated.Metadata["account"]
	if !present {
		return false, "No account data found for: " + validated.Name
	}
	account, isAccount := accountI.(FormiteAccount)
	if !isAccount {
		return false, "Failed to cast account data as Account object for processing: " + validated.Name
	}
	policy, err := account.AccessPolicy()
	if err != nil {
		return false, "Error getting Access Policy for " + validated.Name + ": " + err.Error()
	}
	ok = policy.ContainsTime(time.Now().Local())
	if ok {
		return ok, validated.Name + " validated for this time period."
	}
	return ok, validated.Name + " is not permitted to enter at this time."
}
