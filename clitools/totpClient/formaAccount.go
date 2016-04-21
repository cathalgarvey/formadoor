package main

import "github.com/cathalgarvey/formadoor/timepolicy"

// FormiteAccount represents an account on the Forma Door
type FormiteAccount struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	TimePolicy string `json:"time policy"`
	Secret     string `json:"secret"`
}

// AccessPolicy returns the timepolicy.Policy object represented by the
// TimePolicy property of this account. This can then be queried with
// policy.ContainsTime(time.Now()) to test whether the user is permitted
// access at the present moment.
func (fa FormiteAccount) AccessPolicy() (*timepolicy.Policy, error) {
	return timepolicy.ParsePolicy(fa.TimePolicy)
}
