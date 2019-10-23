package cron

import (
	"fmt"
	"testing"
	"time"
)

func TestAt(t *testing.T) {
	var cases = []struct {
		intention string
		cron      *Cron
		input     string
		want      time.Time
		wantErr   error
	}{
		{
			"simple",
			NewCron(),
			"12:00",
			time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
			nil,
		},
		{
			"invalid pattern",
			NewCron(),
			"98:76",
			NewCron().dayTime,
			fmt.Errorf("parsing time \"98:76\": hour out of range"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			testCase.cron.At(testCase.input)

			failed := false

			if len(testCase.cron.errors) == 0 && testCase.wantErr != nil {
				failed = true
			} else if len(testCase.cron.errors) != 0 && testCase.wantErr == nil {
				failed = true
			} else if len(testCase.cron.errors) > 0 && testCase.cron.errors[0].Error() != testCase.wantErr.Error() {
				failed = true
			} else if testCase.cron.dayTime.String() != testCase.want.String() {
				failed = true
			}

			if failed {
				t.Errorf("At() = (%s, %s), want (%s, %s)", testCase.cron.dayTime, testCase.cron.errors, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestFindMatchingDay(t *testing.T) {
	var cases = []struct {
		intention string
		cron      *Cron
		input     time.Time
		want      time.Time
	}{
		{
			"already good",
			NewCron().Monday().At("12:00"),
			time.Date(2019, 10, 21, 12, 0, 0, 0, time.Local),
			time.Date(2019, 10, 21, 12, 0, 0, 0, time.Local),
		},
		{
			"shift a week",
			NewCron().Saturday().At("12:00"),
			time.Date(2019, 10, 20, 12, 0, 0, 0, time.Local),
			time.Date(2019, 10, 26, 12, 0, 0, 0, time.Local),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := testCase.cron.findMatchingDay(testCase.input); result.String() != testCase.want.String() {
				t.Errorf("findMatchingDay() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestHasError(t *testing.T) {
	var cases = []struct {
		intention string
		cron      *Cron
		input     func(error)
		want      bool
	}{
		{
			"empty cron",
			NewCron(),
			nil,
			true,
		},
		{
			"call of given func",
			NewCron(),
			func(err error) {
				fmt.Println(err)
			},
			true,
		},
		{
			"cron with day config",
			NewCron().Sunday(),
			nil,
			false,
		},
		{
			"cron with day config",
			NewCron().Each(time.Minute),
			nil,
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := testCase.cron.hasError(testCase.input); result != testCase.want {
				t.Errorf("hasError() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}
