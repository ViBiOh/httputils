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

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"go.uber.org/mock/gomock"
)

func TestString(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		cron *Cron
		want string
	}{
		"empty": {
			New().In("UTC"),
			"day: 0000000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"sunday": {
			New().In("UTC").Sunday(),
			"day: 0000001, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"monday": {
			New().In("UTC").Monday(),
			"day: 0000010, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"tuesday": {
			New().In("UTC").Tuesday(),
			"day: 0000100, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"wednesday": {
			New().In("UTC").Wednesday(),
			"day: 0001000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"thursday": {
			New().In("UTC").Thursday(),
			"day: 0010000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"friday": {
			New().In("UTC").Friday(),
			"day: 0100000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"saturday": {
			New().In("UTC").Saturday(),
			"day: 1000000, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"weekdays": {
			New().In("UTC").Weekdays(),
			"day: 0111110, at: 08:00, in: UTC, retry: 0 times every 0s",
		},
		"timezone": {
			New().In("UTC").Monday().At("09:00").In("Europe/Paris"),
			"day: 0000010, at: 09:00, in: Europe/Paris, retry: 0 times every 0s",
		},
		"retry case": {
			New().In("UTC").Each(time.Minute * 10).Retry(time.Minute).MaxRetry(5),
			"each: 10m0s, retry: 5 times every 1m0s",
		},
		"full case": {
			New().In("UTC").Weekdays().At("09:45").In("Europe/Paris").Retry(time.Minute).MaxRetry(5),
			"day: 0111110, at: 09:45, in: Europe/Paris, retry: 5 times every 1m0s, in exclusive mode as `test` with 1m0s timeout",
		},
		"error case": {
			New().In("UTC").Weekdays().At("25:45"),
			"day: 0111110, at: 08:00, in: UTC, retry: 0 times every 0s, error=`parsing time \"25:45\": hour out of range`",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisMock := mocks.NewSemaphore(ctrl)

			if intention == "full case" {
				testCase.cron.Exclusive(redisMock, "test", time.Minute)
			}

			if result := testCase.cron.String(); result != testCase.want {
				t.Errorf("String() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestAt(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		cron    *Cron
		input   string
		want    time.Time
		wantErr error
	}{
		"simple": {
			New(),
			"12:00",
			time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
			nil,
		},
		"invalid pattern": {
			New(),
			"98:76",
			New().dayTime,
			fmt.Errorf("parsing time \"98:76\": hour out of range"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

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
				t.Errorf("At() = (`%s`, `%s`), want (`%s`, `%s`)", testCase.cron.dayTime, testCase.cron.errors, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestIn(t *testing.T) {
	t.Parallel()

	timezone, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Error(err)

		return
	}

	type args struct {
		tz string
	}

	cases := map[string]struct {
		instance *Cron
		args     args
		want     time.Time
		wantErr  error
	}{
		"invalid location": {
			New().At("08:00"),
			args{
				tz: "test",
			},
			time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
			errors.New("unknown time zone test"),
		},
		"converted time": {
			New().At("08:00"),
			args{
				tz: "Europe/Paris",
			},
			time.Date(0, 1, 1, 8, 0, 0, 0, timezone),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			testCase.instance.In(testCase.args.tz)

			failed := false

			if len(testCase.instance.errors) == 0 && testCase.wantErr != nil {
				failed = true
			} else if len(testCase.instance.errors) != 0 && testCase.wantErr == nil {
				failed = true
			} else if len(testCase.instance.errors) > 0 && testCase.instance.errors[0].Error() != testCase.wantErr.Error() {
				failed = true
			} else if testCase.instance.dayTime.String() != testCase.want.String() {
				failed = true
			}

			if failed {
				t.Errorf("In() = (`%s`, `%s`), want (`%s`, `%s`)", testCase.instance.dayTime, testCase.instance.errors, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestFindMatchingDay(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		cron  *Cron
		input time.Time
		want  time.Time
	}{
		"already good": {
			New().Tuesday().At("12:00"),
			time.Date(2019, 10, 22, 12, 0, 0, 0, time.UTC),
			time.Date(2019, 10, 22, 12, 0, 0, 0, time.UTC),
		},
		"shift a week": {
			New().Saturday().At("12:00"),
			time.Date(2019, 10, 20, 12, 0, 0, 0, time.UTC),
			time.Date(2019, 10, 26, 12, 0, 0, 0, time.UTC),
		},
		"next week": {
			New().Weekdays().At("12:00"),
			time.Date(2019, 10, 19, 12, 0, 0, 0, time.UTC),
			time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if result := testCase.cron.findMatchingDay(testCase.input); result.String() != testCase.want.String() {
				t.Errorf("findMatchingDay() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestGetTickerDuration(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		cron  *Cron
		clock GetNow
		input bool
		want  time.Duration
	}{
		"retry": {
			New().Retry(time.Minute),
			time.Now,
			true,
			time.Minute,
		},
		"no retry": {
			New().Each(time.Hour).Retry(time.Minute),
			time.Now,
			false,
			time.Hour,
		},
		"each": {
			New().Each(time.Hour),
			time.Now,
			false,
			time.Hour,
		},
		"same day": {
			New().Days().At("13:00").In("UTC"),
			func() time.Time { return time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC) },
			false,
			time.Hour,
		},
		"one week": {
			New().Monday().At("11:00").In("UTC"),
			func() time.Time { return time.Date(2019, 10, 21, 12, 0, 0, 0, time.UTC) },
			false,
			time.Hour * 167,
		},
		"another timezone": {
			New().Wednesday().At("12:00").In("EST"),
			func() time.Time { return time.Date(2019, 10, 23, 12, 0, 0, 0, time.UTC) },
			false,
			time.Hour * 5,
		},
		"hour shift": {
			New().Sunday().At("12:00").In("Europe/Paris"),
			func() time.Time { return time.Date(2019, 10, 26, 22, 0, 0, 0, time.UTC) },
			false,
			time.Hour * 13,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			testCase.cron.clock = testCase.clock
			result := testCase.cron.getTickerDuration(testCase.input)
			if !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("getTickerDuration() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestHasError(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		cron *Cron
		want bool
	}{
		"empty cron": {
			New(),
			true,
		},
		"empty with invalid value": {
			New().At("98:76"),
			true,
		},
		"empty with invalid timezone": {
			New().In("Rainbow"),
			true,
		},
		"days and interval": {
			New().Monday().Each(time.Minute),
			true,
		},
		"retry without interval": {
			New().Weekdays().MaxRetry(5),
			true,
		},
		"cron with day config": {
			New().Friday(),
			false,
		},
		"cron with interval config": {
			New().Each(time.Minute),
			false,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if result := testCase.cron.hasError(context.Background()); result != testCase.want {
				t.Errorf("hasError() = %t, want %t", result, testCase.want)
			}
		})
	}
}

func TestStart(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		cron    *Cron
		clock   GetNow
		action  func(*sync.WaitGroup, *Cron) func(context.Context) error
		onError func(*sync.WaitGroup, *Cron) func(context.Context, error)
	}{
		"run once": {
			New().Days().At("12:00"),
			func() time.Time { return time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC) },
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				return func(_ context.Context) error {
					wg.Done()

					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context, err error) {
				return func(_ context.Context, err error) {
					t.Error(errors.New("should not be there"))
				}
			},
		},
		"retry": {
			New().Days().At("12:00").Retry(time.Millisecond).MaxRetry(5),
			func() time.Time { return time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC) },
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
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context, err error) {
				return func(_ context.Context, err error) {}
			},
		},
		"run on demand": {
			New().Days().At("12:00"),
			func() time.Time { return time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC) },
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				cron.Now()

				return func(_ context.Context) error {
					wg.Done()

					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context, err error) {
				return func(_ context.Context, err error) {
					t.Error(fmt.Errorf("should not be there: %w", err))
				}
			},
		},
		"run on signal": {
			New().Days().At("12:00").OnSignal(syscall.SIGUSR1),
			func() time.Time { return time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC) },
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
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context, err error) {
				return func(_ context.Context, err error) {
					t.Error(fmt.Errorf("should not be there: %w", err))
				}
			},
		},
		"run in exclusive error": {
			New().Days().At("12:00"),
			func() time.Time { return time.Date(2019, 10, 21, 11, 59, 59, 900, time.UTC) },
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				return func(_ context.Context) error {
					t.Error(errors.New("should not be there"))

					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context, err error) {
				return func(_ context.Context, err error) {
					wg.Done()
				}
			},
		},
		"fail if misconfigured": {
			New(),
			func() time.Time { return time.Date(2019, 10, 21, 11, 0, 0, 0, time.UTC) },
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context) error {
				cron.Now()

				return func(_ context.Context) error {
					wg.Done()

					return nil
				}
			},
			func(wg *sync.WaitGroup, cron *Cron) func(_ context.Context, err error) {
				return func(_ context.Context, err error) {
					wg.Done()
				}
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisMock := mocks.NewSemaphore(ctrl)

			if intention == "run in exclusive error" {
				testCase.cron.Exclusive(redisMock, "test", time.Minute)
				redisMock.EXPECT().Exclusive(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, errors.New("redis error"))
			}

			var wg sync.WaitGroup
			wg.Add(1)
			testCase.cron.clock = testCase.clock

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go testCase.cron.OnError(testCase.onError(&wg, testCase.cron)).Start(ctx, testCase.action(&wg, testCase.cron))

			actionDone := make(chan struct{})
			go func() {
				wg.Wait()
				close(actionDone)
			}()

			select {
			case <-time.After(time.Second * 5):
				t.Errorf("Start() did not complete within 5 seconds")
			case <-actionDone:
			}
		})
	}
}
