package timepolicy

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidPolicyBoundString is returned if the general form of a policy string
	// is bad, and not a more specific part.
	ErrInvalidPolicyBoundString = errors.New("Structure of PolicyBound string spec is bad")

	// ErrInvalidClockTimeString is returned on bad "HH:MM->HH:MM" specs.
	ErrInvalidClockTimeString = errors.New("Bad ClockTime spec: must be `HH:MM->HH:MM`")

	// ErrInvalidDayString is returned on bad "[DOW:DOW]" specs.
	ErrInvalidDayString = errors.New("Bad day of week spec; must be `[DOW:DOW]`")

	// ErrMismatchedClockTimes is returned if format is correct but times are out of order.
	ErrMismatchedClockTimes = errors.New("Formatted time has times in wrong order")
)

// Policies take the form: [N:N]HH:MM->HH:MM, with optional bar-separated repeats
// containing additional PolicyBounds
// In the original version weekdays were zero-indexed, now instead they are
// three-char weekday abbreviations, for eg: [Mon:Sun]08:30->23:59

// ParsePolicyBound accepts a single string of form [DoW:DoW]HH:MM->HH:MM and
// returns a PolicyBound. It errors if the string appears malformed (whitespace
// is trimmed and ignored) or if the Bound makes no sense.
func ParsePolicyBound(bound string) (*PolicyBound, error) {
	bound = strings.TrimSpace(bound)
	dayBoundary := strings.Index(bound, "]")
	// Must have "]", must be long enough to have a valid timebound section.
	if dayBoundary == -1 || dayBoundary > (len(bound)-12) {
		return nil, ErrInvalidPolicyBoundString
	}
	days, err := parseDayBits(bound[:dayBoundary+1])
	if err != nil {
		return nil, err
	}
	lowerTime, upperTime, err := parseTimeBits(bound[dayBoundary+1:])
	if err != nil {
		return nil, err
	}
	return NewPolicyBound(*lowerTime, *upperTime, days...)
}

func parseDayBits(daybits string) ([]time.Weekday, error) {
	daybits = strings.TrimSpace(daybits)
	if (!strings.HasPrefix(daybits, "[")) || (!strings.HasSuffix(daybits, "]")) {
		return nil, ErrInvalidDayString
	}
	daybits = strings.Trim(daybits, "[]")
	hldays := strings.Split(daybits, ":")
	if len(hldays) != 2 {
		return nil, ErrInvalidDayString
	}
	lowDay, err := dowToWeekday(hldays[0])
	if err != nil {
		return nil, err
	}
	highDay, err := dowToWeekday(hldays[1])
	if err != nil {
		return nil, err
	}
	return getDayRange(lowDay, highDay), nil
}

func dowToWeekday(daybit string) (time.Weekday, error) {
	daybit = strings.TrimSpace(strings.ToLower(daybit))
	if len(daybit) < 3 {
		return -1, ErrInvalidDayString
	}
	switch daybit[:3] {
	case "mon":
		return time.Monday, nil
	case "tue":
		return time.Tuesday, nil
	case "wed":
		return time.Wednesday, nil
	case "thu":
		return time.Thursday, nil
	case "fri":
		return time.Friday, nil
	case "sat":
		return time.Saturday, nil
	case "sun":
		return time.Sunday, nil
	default:
		return -1, ErrInvalidDayString
	}
}

func getDayRange(lowDay, highDay time.Weekday) []time.Weekday {
	var days []time.Weekday
	if lowDay < 0 || lowDay > 6 || highDay < 0 || highDay > 6 {
		panic("lowDay or highDay fall outside permitted range!")
	}
	if lowDay == highDay {
		days = []time.Weekday{lowDay}
		return days
	}
	if lowDay < highDay {
		for i := lowDay; i <= highDay; i++ {
			days = append(days, i)
		}
		return days
	}
	if lowDay > highDay {
		for i := lowDay; i <= time.Saturday; i++ {
			days = append(days, i)
		}
		for i := time.Sunday; i <= highDay; i++ {
			days = append(days, i)
		}
		return days
	}
	return nil
}

func parseTimeBits(timebits string) (low, high *ClockTime, err error) {
	timebits = strings.TrimSpace(strings.ToLower(timebits))
	bits := strings.Split(timebits, "->")
	if len(bits) != 2 {
		return nil, nil, ErrInvalidClockTimeString
	}
	lowBit, err := timeStrToClockTime(bits[0])
	if err != nil {
		return nil, nil, err
	}
	highBit, err := timeStrToClockTime(bits[1])
	if err != nil {
		return nil, nil, err
	}
	lowToday, err := lowBit.toTimeToday()
	if err != nil {
		return nil, nil, err
	}
	highToday, err := highBit.toTimeToday()
	if err != nil {
		return nil, nil, err
	}
	if lowToday.After(highToday) {
		return nil, nil, ErrMismatchedClockTimes
	}
	return lowBit, highBit, nil
}

func timeStrToClockTime(timestr string) (*ClockTime, error) {
	if len(timestr) < len("HH:MM") {
		return nil, ErrInvalidClockTimeString
	}
	bits := strings.Split(timestr, ":")
	if len(bits) != 2 {
		return nil, ErrInvalidClockTimeString
	}
	hourS := strings.TrimSpace(strings.ToLower(bits[0]))
	minS := strings.TrimSpace(strings.ToLower(bits[1]))
	if len(hourS) != 2 || len(minS) != 2 {
		return nil, ErrInvalidClockTimeString
	}
	hour, err := strconv.Atoi(hourS)
	if err != nil {
		return nil, ErrInvalidClockTimeString
	}
	min, err := strconv.Atoi(minS)
	if err != nil {
		return nil, ErrInvalidClockTimeString
	}
	ct := ClockTime{Hour: hour, Minute: min}
	if !ct.isValid() {
		return nil, ErrInvalidClockTimeString
	}
	return &ct, nil
}
