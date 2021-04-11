package cron

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	var cases = []struct {
		intention string
		cron      *Cron
		want      string
	}{
		{
			"empty",
			New().In("UTC"),
			"day: 0000000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"sunday",
			New().In("UTC").Sunday(),
			"day: 0000001, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"monday",
			New().In("UTC").Monday(),
			"day: 0000010, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"tuesday",
			New().In("UTC").Tuesday(),
			"day: 0000100, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"wednesday",
			New().In("UTC").Wednesday(),
			"day: 0001000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"thursday",
			New().In("UTC").Thursday(),
			"day: 0010000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"friday",
			New().In("UTC").Friday(),
			"day: 0100000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"saturday",
			New().In("UTC").Saturday(),
			"day: 1000000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"weekdays",
			New().In("UTC").Weekdays(),
			"day: 0111110, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		{
			"timezone",
			New().In("UTC").Monday().At("09:00").In("Europe/Paris"),
			"day: 0000010, at: 09:00, in: Europe/Paris, retry: 0 times every 0s",
		},
		{
			"retry case",
			New().In("UTC").Each(time.Minute * 10).Retry(time.Minute).MaxRetry(5),
			"each: 10m0s, retry: 5 times every 1m0s",
		},
		{
			"full case",
			New().In("UTC").Weekdays().At("09:45").In("Europe/Paris").Retry(time.Minute).MaxRetry(5),
			"day: 0111110, at: 09:45, in: Europe/Paris, retry: 5 times every 1m0s",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if result := tc.cron.String(); result != tc.want {
				t.Errorf("String() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			tc.cron.At(tc.input)

			failed := false

			if len(tc.cron.errors) == 0 && tc.wantErr != nil {
				failed = true
			} else if len(tc.cron.errors) != 0 && tc.wantErr == nil {
				failed = true
			} else if len(tc.cron.errors) > 0 && tc.cron.errors[0].Error() != tc.wantErr.Error() {
				failed = true
			} else if tc.cron.dayTime.String() != tc.want.String() {
				failed = true
			}

			if failed {
				t.Errorf("At() = (`%s`, `%s`), want (`%s`, `%s`)", tc.cron.dayTime, tc.cron.errors, tc.want, tc.wantErr)
			}
		})
	}
}

func TestIn(t *testing.T) {
	timezone, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		tz string
	}

	var cases = []struct {
		intention string
		instance  *Cron
		args      args
		want      time.Time
		wantErr   error
	}{
		{
			"invalid location",
			New().At("08:00"),
			args{
				tz: "test",
			},
			time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
			errors.New("unknown time zone test"),
		},
		{
			"converted time",
			New().At("08:00"),
			args{
				tz: "Europe/Paris",
			},
			time.Date(0, 1, 1, 8, 0, 0, 0, timezone),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			tc.instance.In(tc.args.tz)

			failed := false

			if len(tc.instance.errors) == 0 && tc.wantErr != nil {
				failed = true
			} else if len(tc.instance.errors) != 0 && tc.wantErr == nil {
				failed = true
			} else if len(tc.instance.errors) > 0 && tc.instance.errors[0].Error() != tc.wantErr.Error() {
				failed = true
			} else if tc.instance.dayTime.String() != tc.want.String() {
				failed = true
			}

			if failed {
				t.Errorf("In() = (`%s`, `%s`), want (`%s`, `%s`)", tc.instance.dayTime, tc.instance.errors, tc.want, tc.wantErr)
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
			time.Date(2019, 10, 22, 12, 0, 0, 0, time.UTC),
			time.Date(2019, 10, 22, 12, 0, 0, 0, time.UTC),
		},
		{
			"shift a week",
			New().Saturday().At("12:00"),
			time.Date(2019, 10, 20, 12, 0, 0, 0, time.UTC),
			time.Date(2019, 10, 26, 12, 0, 0, 0, time.UTC),
		},
		{
			"next week",
			New().Weekdays().At("12:00"),
			time.Date(2019, 10, 19, 12, 0, 0, 0, time.UTC),
			time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if result := tc.cron.findMatchingDay(tc.input); result.String() != tc.want.String() {
				t.Errorf("findMatchingDay() = `%s`, want `%s`", result, tc.want)
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
		{
			"hour shift",
			New().Sunday().At("12:00").In("Europe/Paris").Clock(&Clock{time.Date(2019, 10, 26, 22, 0, 0, 0, time.UTC)}),
			false,
			time.Hour * 13,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			result := tc.cron.getTickerDuration(tc.input)
			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("getTickerDuration() = `%s`, want `%s`", result, tc.want)
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
			"days and interval",
			New().Monday().Each(time.Minute),
			true,
		},
		{
			"retry without interval",
			New().Weekdays().MaxRetry(5),
			true,
		},
		{
			"cron with day config",
			New().Friday(),
			false,
		},
		{
			"cron with day config",
			New().Each(time.Minute),
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if result := tc.cron.hasError(); result != tc.want {
				t.Errorf("hasError() = %t, want %t", result, tc.want)
			}
		})
	}
}

func TestStart(t *testing.T) {
	var cases = []struct {
		intention string
		cron      *Cron
		action    func(*sync.WaitGroup, *Cron) func(context.Context) error
		onError   func(*sync.WaitGroup, *Cron) func(error)
	}{
		{
			"run once",
			New().Days().At("12:00").Clock(&Clock{time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC)}),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				return func(_ context.Context) error {
					wg.Done()

					cron.Clock(&Clock{time.Date(2019, 10, 21, 13, 0, 0, 0, time.UTC)})
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {
					t.Error(err)
				}
			},
		},
		{
			"retry",
			New().Days().At("12:00").Retry(time.Millisecond).MaxRetry(5).Clock(&Clock{time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC)}),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				count := 0
				return func(_ context.Context) error {
					count++
					if count < 4 {
						return errors.New("call me again")
					}

					wg.Done()
					cron.Clock(&Clock{time.Date(2019, 10, 21, 13, 0, 0, 0, time.UTC)})
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {}
			},
		},
		{
			"run on demand",
			New().Days().At("12:00").Clock(&Clock{time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC)}),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				cron.Now()

				return func(_ context.Context) error {
					wg.Done()
					cron.Clock(&Clock{time.Date(2019, 10, 21, 13, 0, 0, 0, time.UTC)})
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {}
			},
		},
		{
			"run on signal",
			New().Days().At("12:00").Clock(&Clock{time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC)}).OnSignal(syscall.SIGUSR1),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				p, err := os.FindProcess(os.Getpid())
				if err != nil {
					t.Error(err)
				}

				go func() {
					time.Sleep(time.Second)
					p.Signal(syscall.SIGUSR1)
				}()

				return func(_ context.Context) error {
					wg.Done()
					cron.Clock(&Clock{time.Date(2019, 10, 21, 13, 0, 0, 0, time.UTC)})
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {}
			},
		},
		{
			"fail if misconfigured",
			New().Clock(&Clock{time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC)}),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				cron.Now()

				return func(_ context.Context) error {
					wg.Done()
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {
					wg.Done()
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)

			go tc.cron.OnError(tc.onError(&wg, tc.cron)).Start(tc.action(&wg, tc.cron), nil)

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-time.After(time.Second * 5):
				tc.cron.Shutdown()
				t.Errorf("Start() did not complete within 5 seconds")
			case <-done:
				tc.cron.Shutdown()
			}
		})
	}
}
