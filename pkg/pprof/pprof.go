package pprof

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	_ "net/http/pprof"
	"runtime/pprof"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var ProfileNames = []string{
	"allocs",
	"goroutine",
	"heap",
}

var cpuDuration = time.Second * 30

type Service struct {
	buffer  *bytes.Buffer
	service string
	version string
	env     string
	req     request.Request
	port    int
}

type Config struct {
	URL  string
	Port int
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Agent", "URL of the Datadog Trace Agent (e.g. http://datadog.observability:8126)").Prefix(prefix).DocPrefix("pprof").StringVar(fs, &config.URL, "", overrides)
	flags.New("Port", "Port of the HTTP server (0 to disable)").Prefix(prefix).DocPrefix("pprof").IntVar(fs, &config.Port, 0, overrides)

	return &config
}

func New(config *Config, service, version, env string) *Service {
	instance := Service{
		port: config.Port,
	}

	if len(config.URL) != 0 {
		instance.req = request.Post(fmt.Sprintf("%s/profiling/v1/input", config.URL)).WithClient(request.CreateClient(15*time.Second, request.NoRedirection))
		instance.buffer = bytes.NewBuffer(nil)
		instance.service = service
		instance.version = version
		instance.env = env
	}

	return &instance
}

func (s *Service) Start(ctx context.Context) {
	if s.port > 0 {
		go s.http(ctx)
	}

	if s.buffer != nil {
		go s.push(ctx)
	}
}

func (s *Service) http(ctx context.Context) {
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", s.port), http.DefaultServeMux); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("pprof http: %s", err.Error()))
	}
}

func (s *Service) push(ctx context.Context) {
	defer recoverer.Logger()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.send(ctx); !httperror.CanBeIgnored(err) {
				slog.LogAttrs(ctx, slog.LevelError, "pprof export", slog.Any("error", err))
			}
		}
	}
}

func (s *Service) send(ctx context.Context) error {
	now := time.Now()

	if err := s.getCpuProfile(); err != nil {
		return fmt.Errorf("get cpu profile: %w", err)
	}

	resp, err := s.req.Multipart(ctx, s.writeMultipart(ctx, now))
	if err != nil {
		return fmt.Errorf("send multipart: %w", err)
	}

	if err := request.DiscardBody(resp.Body); err != nil {
		return fmt.Errorf("discard body: %w", err)
	}

	return nil
}

func (s *Service) getCpuProfile() error {
	s.buffer.Reset()

	if err := pprof.StartCPUProfile(s.buffer); err != nil {
		return fmt.Errorf("start profiler: %w", err)
	}

	time.Sleep(cpuDuration)
	pprof.StopCPUProfile()

	return nil
}

func (s *Service) writeMultipart(ctx context.Context, now time.Time) func(*multipart.Writer) error {
	return func(mw *multipart.Writer) error {
		if err := mw.WriteField("version", "3"); err != nil {
			return fmt.Errorf("write field `version`: %w", err)
		}

		if err := mw.WriteField("format", "pprof"); err != nil {
			return fmt.Errorf("write field `format`: %w", err)
		}

		if err := mw.WriteField("family", "go"); err != nil {
			return fmt.Errorf("write field `family`: %w", err)
		}

		if err := mw.WriteField("start", now.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("write field `start`: %w", err)
		}

		if err := mw.WriteField("end", now.Add(cpuDuration).Format(time.RFC3339)); err != nil {
			return fmt.Errorf("write field `end`: %w", err)
		}

		if err := mw.WriteField("tags[]", "runtime:go"); err != nil {
			return fmt.Errorf("write field `tags` for `runtime`: %w", err)
		}

		if err := mw.WriteField("tags[]", fmt.Sprintf("service:%s", s.service)); err != nil {
			return fmt.Errorf("write field `tags` for `service`: %w", err)
		}

		if err := mw.WriteField("tags[]", fmt.Sprintf("version:%s", s.version)); err != nil {
			return fmt.Errorf("write field `tags` for `version`: %w", err)
		}

		if err := mw.WriteField("tags[]", fmt.Sprintf("env:%s", s.env)); err != nil {
			return fmt.Errorf("write field `tags` for `env`: %w", err)
		}

		if err := addCPU(mw, s.buffer); err != nil {
			return fmt.Errorf("add profile `cpu`: %w", err)
		}

		for _, name := range ProfileNames {
			profile := pprof.Lookup(name)
			if profile == nil {
				slog.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("unknown profile `%s`", name))
				continue
			}

			if err := addProfile(mw, profile); err != nil {
				return fmt.Errorf("add profile `%s`: %w", profile.Name(), err)
			}
		}

		return nil
	}
}

func addCPU(mw *multipart.Writer, buffer *bytes.Buffer) error {
	partWriter, err := mw.CreateFormFile("data[cpu.pprof]", "cpu.pprof")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if _, err := buffer.WriteTo(partWriter); err != nil {
		return fmt.Errorf("write profile: %w", err)
	}

	return nil
}

func addProfile(mw *multipart.Writer, profile *pprof.Profile) error {
	partWriter, err := mw.CreateFormFile(fmt.Sprintf("data[%s.pprof]", profile.Name()), profile.Name())
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if err = profile.WriteTo(partWriter, 0); err != nil {
		return fmt.Errorf("write profile: %w", err)
	}

	return nil
}
