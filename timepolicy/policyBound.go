package timepolicy

import (
	"errors"
	"time"
)

var (
	// ErrInvalidClockTime is returned if hours or minutes in a ClockTime struct fall
	// outside expected 24h bounds.
	ErrInvalidClockTime = errors.New("Invalid clock time, cannot use values meaningfully")
)

// ClockTime is a limited view on time that's only concerned with clock/watch-time,
// eg. Hours:Minutes, in 24-hour notation.
type ClockTime struct {
	Hour   int
	Minute int
}

func (ct ClockTime) toTimeToday() (time.Time, error) {
	if !ct.isValid() {
		return time.Time{}, ErrInvalidClockTime
	}
	now := time.Now().Local()
	return time.Date(now.Year(), now.Month(), now.Day(), ct.Hour, ct.Minute, 0, 0, now.Location()), nil
}

func (ct ClockTime) isValid() bool {
	if ct.Hour < 0 || ct.Hour > 23 {
		return false
	}
	if ct.Minute < 0 || ct.Minute > 59 {
		return false
	}
	return true
}

// PolicyBound is a policy describing hours of access on a set of days.
// When asked for validity of a timepoint, PolicyBound checks whether that
// timepoint's day is within Days, and if so, whether that timepoint's
// clock-time is within the LowerTime->UpperTime frame.
type PolicyBound struct {
	Days      []time.Weekday
	LowerTime ClockTime
	UpperTime ClockTime
}

// NewPolicyBound is a shortcut for creating PolicyBound directly that also
// validates the ClockTimes and offers a simpler variadic way to write Days.
func NewPolicyBound(LowerT, UpperT ClockTime, Days ...time.Weekday) (*PolicyBound, error) {
	if (!LowerT.isValid()) || (!UpperT.isValid()) {
		return nil, ErrInvalidClockTime
	}
	return &PolicyBound{
		Days:      Days,
		LowerTime: LowerT,
		UpperTime: UpperT,
	}, nil
}

// ContainsTime checks whether a time lies within the PolicyBound, eg. whether
// the weekdayof this time is a weekday permitted by the policy, and then
// whether the time of day is valid within the present day.
func (pb PolicyBound) ContainsTime(t time.Time) bool {
	tloc := t.Local()
	td := tloc.Weekday()
	for _, day := range pb.Days {
		if td == day {
			goto okday
		}
	}
	return false
okday: // Yolo
	tl, err := pb.LowerTime.toTimeToday()
	if err != nil {
		return false
	}
	tu, err := pb.UpperTime.toTimeToday()
	if err != nil {
		return false
	}
	if t.Before(tl) {
		return false
	}
	if tu.Before(t) {
		return false
	}
	return true
}
