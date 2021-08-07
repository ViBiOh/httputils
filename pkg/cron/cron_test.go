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

	"github.com/ViBiOh/httputils/v4/pkg/clock"
	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/golang/mock/gomock"
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
			"day: 0111110, at: 09:45, in: Europe/Paris, retry: 5 times every 1m0s, in exclusive mode as `test` with 1m0s timeout",
		},
		{
			"error case",
			New().In("UTC").Weekdays().At("25:45"),
			"day: 0111110, at: 08:00, in: UTC, retry: 0 times every 0s, error=`parsing time \"25:45\": hour out of range`",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisMock := mocks.NewRedis(ctrl)

			if tc.intention == "full case" {
				tc.cron.Exclusive(redisMock, "test", time.Minute)
			}

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
		clock     *clock.Clock
		input     bool
		want      time.Duration
	}{
		{
			"retry",
			New().Retry(time.Minute),
			nil,
			true,
			time.Minute,
		},
		{
			"no retry",
			New().Each(time.Hour).Retry(time.Minute),
			nil,
			false,
			time.Hour,
		},
		{
			"each",
			New().Each(time.Hour),
			nil,
			false,
			time.Hour,
		},
		{
			"same day",
			New().Days().At("13:00").In("UTC"),
			clock.New(time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC)),
			false,
			time.Hour,
		},
		{
			"one week",
			New().Monday().At("11:00").In("UTC"),
			clock.New(time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC)),
			false,
			time.Hour * 167,
		},
		{
			"another timezone",
			New().Wednesday().At("12:00").In("EST"),
			clock.New(time.Date(2019, 10, 23, 12, 0, 0, 0, time.UTC)),
			false,
			time.Hour * 5,
		},
		{
			"hour shift",
			New().Sunday().At("12:00").In("Europe/Paris"),
			clock.New(time.Date(2019, 10, 26, 22, 0, 0, 0, time.UTC)),
			false,
			time.Hour * 13,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			tc.cron.clock = tc.clock
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
		clock     *clock.Clock
		action    func(*sync.WaitGroup, *Cron) func(context.Context) error
		onError   func(*sync.WaitGroup, *Cron) func(error)
	}{
		{
			"run once",
			New().Days().At("12:00"),
			clock.New(time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC)),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				return func(_ context.Context) error {
					wg.Done()

					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {
					t.Error(errors.New("should not be there"))
				}
			},
		},
		{
			"retry",
			New().Days().At("12:00").Retry(time.Millisecond).MaxRetry(5),
			clock.New(time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC)),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				count := 0
				return func(_ context.Context) error {
					count++
					if count < 4 {
						return errors.New("call me again")
					}

					wg.Done()
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {}
			},
		},
		{
			"run on demand",
			New().Days().At("12:00"),
			clock.New(time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC)),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				cron.Now()

				return func(_ context.Context) error {
					wg.Done()
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {
					t.Error(fmt.Errorf("should not be there: %s", err))
				}
			},
		},
		{
			"run on signal",
			New().Days().At("12:00").OnSignal(syscall.SIGUSR1),
			clock.New(time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC)),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				p, err := os.FindProcess(os.Getpid())
				if err != nil {
					t.Error(err)
				}

				go func() {
					time.Sleep(time.Second)
					if err := p.Signal(syscall.SIGUSR1); err != nil {
						fmt.Println(err)
					}
				}()

				return func(_ context.Context) error {
					wg.Done()
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {
					t.Error(fmt.Errorf("should not be there: %s", err))
				}
			},
		},
		{
			"run in exclusive error",
			New().Days().At("12:00"),
			clock.New(time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC)),
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				return func(_ context.Context) error {
					t.Error(errors.New("should not be there"))
					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(err error) {
				return func(err error) {
					wg.Done()
				}
			},
		},
		{
			"fail if misconfigured",
			New(),
			clock.New(time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC)),
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisMock := mocks.NewRedis(ctrl)

			if tc.intention == "run in exclusive error" {
				tc.cron.Exclusive(redisMock, "test", time.Minute)
				redisMock.EXPECT().Exclusive(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("redis error"))
			}

			var wg sync.WaitGroup
			wg.Add(1)
			tc.cron.clock = tc.clock

			go tc.cron.OnError(tc.onError(&wg, tc.cron)).Start(tc.action(&wg, tc.cron), nil)

			actionDone := make(chan struct{})
			go func() {
				wg.Wait()
				close(actionDone)
			}()

			select {
			case <-time.After(time.Second * 5):
				tc.cron.Shutdown()
				t.Errorf("Start() did not complete within 5 seconds")
			case <-actionDone:
				tc.cron.Shutdown()
			}
		})
	}
}
