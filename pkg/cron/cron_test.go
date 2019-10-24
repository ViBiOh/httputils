package cron

import (
	"fmt"
	"reflect"
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
			New(),
			"12:00",
			time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
			nil,
		},
		{
			"invalid pattern",
			New(),
			"98:76",
			New().dayTime,
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
			New().Tuesday().At("12:00"),
			time.Date(2019, 10, 22, 12, 0, 0, 0, time.Local),
			time.Date(2019, 10, 22, 12, 0, 0, 0, time.Local),
		},
		{
			"shift a week",
			New().Saturday().At("12:00"),
			time.Date(2019, 10, 20, 12, 0, 0, 0, time.Local),
			time.Date(2019, 10, 26, 12, 0, 0, 0, time.Local),
		},
		{
			"next week",
			New().Weekdays().At("12:00"),
			time.Date(2019, 10, 19, 12, 0, 0, 0, time.Local),
			time.Date(2019, 10, 21, 12, 0, 0, 0, time.Local),
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

func TestGetTickerDuration(t *testing.T) {
	var cases = []struct {
		intention string
		cron      *Cron
		input     bool
		want      time.Duration
	}{
		{
			"retry",
			New().Retry(time.Minute),
			true,
			time.Minute,
		},
		{
			"no retry",
			New().Each(time.Hour).Retry(time.Minute),
			false,
			time.Hour,
		},
		{
			"each",
			New().Each(time.Hour),
			false,
			time.Hour,
		},
		{
			"same day",
			New().Days().At("13:00").In("UTC").Clock(&Clock{time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC)}),
			false,
			time.Hour,
		},
		{
			"one week",
			New().Monday().At("11:00").In("UTC").Clock(&Clock{time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC)}),
			false,
			time.Hour * 167,
		},
		{
			"another timezone",
			New().Wednesday().At("12:00").In("EST").Clock(&Clock{time.Date(2019, 10, 23, 12, 0, 0, 0, time.UTC)}),
			false,
			time.Hour * 5,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result := testCase.cron.getTickerDuration(testCase.input)
			if !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("getTickerDuration() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestHasError(t *testing.T) {
	var cases = []struct {
		intention string
		cron      *Cron
		want      bool
	}{
		{
			"empty cron",
			New(),
			true,
		},
		{
			"empty with invalid value",
			New().At("98:76"),
			true,
		},
		{
			"empty with invalid timezone",
			New().In("Rainbow"),
			true,
		},
		{
			"cron with day config",
			New().Sunday(),
			false,
		},
		{
			"cron with day config",
			New().Each(time.Minute),
			false,
		},
	}

	onError := func(err error) {
		fmt.Println(err)
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := testCase.cron.hasError(onError); result != testCase.want {
				t.Errorf("hasError() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}
