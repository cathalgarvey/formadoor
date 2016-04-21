package timepolicy

import (
	"strings"
	"time"
)

// Policy is a set of PolicyBounds, any of which can validate. Overlaps are
// irrelevant.
type Policy struct {
	Bounds []PolicyBound
}

// ContainsTime checks whether a time is within any contained PolicyBound.
func (p Policy) ContainsTime(t time.Time) bool {
	for _, pb := range p.Bounds {
		if pb.ContainsTime(t) {
			return true
		}
	}
	return false
}

// ParsePolicy takes a string of form `[dow:dow]hh:mm->hh:mm|[dow:dow]hh:mm->hh:mm...`
// and creates a policy.
func ParsePolicy(policyString string) (*Policy, error) {
	policy := new(Policy)
	PBStrings := strings.Split(policyString, "|")
	for _, pbs := range PBStrings {
		parsedPolicy, err := ParsePolicyBound(pbs)
		if err != nil {
			return nil, err
		}
		policy.Bounds = append(policy.Bounds, *parsedPolicy)
	}
	return policy, nil
}
