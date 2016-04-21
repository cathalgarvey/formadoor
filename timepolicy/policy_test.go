package timepolicy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPolicyParsing(t *testing.T) {

}

func TestTimePolicyBoundsParsing(t *testing.T) {
	testT, err := ParsePolicyBound("[mon:friday]08:00->20:30")
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	assert.EqualValues(t, []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}, testT.Days)
	assert.EqualValues(t, 8, testT.LowerTime.Hour)
	assert.EqualValues(t, 0, testT.LowerTime.Minute)
	assert.EqualValues(t, 20, testT.UpperTime.Hour)
	assert.EqualValues(t, 30, testT.UpperTime.Minute)

	validDate := time.Date(2016, time.April, 8, 9, 37, 0, 0, time.Local)
	invalidDate := time.Date(2016, time.April, 9, 9, 37, 0, 0, time.Local)
	assert.True(t, testT.ContainsTime(validDate))
	assert.False(t, testT.ContainsTime(invalidDate))
}

func TestClockTimeStringParsing(t *testing.T) {
	testT, err := timeStrToClockTime("12 : 12")
	assert.Nil(t, err)
	assert.EqualValues(t, 12, testT.Hour)
	assert.EqualValues(t, 12, testT.Minute)

	testT, err = timeStrToClockTime("12 : 59")
	assert.Nil(t, err)
	assert.EqualValues(t, 12, testT.Hour)
	assert.EqualValues(t, 59, testT.Minute)

	testT, err = timeStrToClockTime("12:00")
	assert.Nil(t, err)
	assert.EqualValues(t, 12, testT.Hour)
	assert.EqualValues(t, 00, testT.Minute)

	lowT, highT, err := parseTimeBits("08:00->20:30")
	assert.Nil(t, err)
	assert.EqualValues(t, 8, lowT.Hour)
	assert.EqualValues(t, 0, lowT.Minute)
	assert.EqualValues(t, 20, highT.Hour)
	assert.EqualValues(t, 30, highT.Minute)

	testT, err = timeStrToClockTime("23 : -1")
	assert.Error(t, err)

	testT, err = timeStrToClockTime("24 : 00")
	assert.Error(t, err)

	testT, err = timeStrToClockTime("23:61")
	assert.Error(t, err)

	testT, err = timeStrToClockTime("23 : -1")
	assert.Error(t, err)

	testT, err = timeStrToClockTime("23:")
	assert.Error(t, err)
	testT, err = timeStrToClockTime("23:0")
	assert.Error(t, err)
	testT, err = timeStrToClockTime("1:00")
	assert.Error(t, err)
}

func TestTimePolicyBoundsString(t *testing.T) {
	mon, err := dowToWeekday("Mon")
	assert.Nil(t, err)
	wed, err := dowToWeekday("wed")
	assert.Nil(t, err)
	testDays := getDayRange(mon, wed)
	assert.EqualValues(t, []time.Weekday{time.Monday, time.Tuesday, time.Wednesday}, testDays)

	foo, err := dowToWeekday("fooday")
	assert.Equal(t, time.Weekday(-1), foo)
	assert.EqualError(t, err, ErrInvalidDayString.Error())

	sat, err := dowToWeekday("sat")
	assert.Nil(t, err)
	sun, err := dowToWeekday("sunday")
	assert.Nil(t, err)
	testDays = getDayRange(sat, sun)
	assert.EqualValues(t, []time.Weekday{time.Saturday, time.Sunday}, testDays)

	testDays = getDayRange(sat, sat)
	assert.EqualValues(t, []time.Weekday{time.Saturday}, testDays)

	testDays = getDayRange(wed, mon)
	assert.EqualValues(t, []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday}, testDays)
}
