# httputils

[![Build](https://github.com/ViBiOh/httputils/workflows/Build/badge.svg)](https://github.com/ViBiOh/httputils/actions)

Golang HTTP Utils func

## Alcotest

Make your server blow into balloon for checking its health.

### Usage

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of alcotest:
  -url string
        [alcotest] URL to check {ALCOTEST_URL}
  -userAgent string
        [alcotest] User-Agent for check {ALCOTEST_USER_AGENT} (default "Alcotest")
```

in `Dockerfile`

```bash
HEALTHCHECK --retries=10 CMD /alcotest -url http://localhost/health

COPY bin/alcotest /alcotest
```

### Why `alcotest` ?

Because main use is to healthcheck.
Because it's a partial anagram of "ch**e**ck **stat**us **co**de ur**l**"

## Server

Server is provided in order to be compliant with [12 Factor](https://12factor.net).

### Middlewares

#### health

Handle healthcheck status of the server.

#### recoverer

Catch `panic`, respond `HTTP/500` and log the beginning of the stacktrace.

#### owasp / cors

Enforce security best practices for serving web content.

### Endpoints

- `GET /health`: healthcheck of server, always respond [`okStatus (default 204)`](#Usage)
- `GET /ready`: checks external dependencies availability and then respond [`okStatus (default 204)`](#usage) or `503` during [`graceDuration`](#usage) when close signal is received
- `GET /version`: value of `VERSION` environment variable

### Usage

App can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of http:
  --address              string        [server] Listen address ${HTTP_ADDRESS}
  --amqpExchange         string        [amqp] Exchange name ${HTTP_AMQP_EXCHANGE} (default "httputils")
  --amqpExclusive                      [amqp] Queue exclusive mode (for fanout exchange) ${HTTP_AMQP_EXCLUSIVE} (default false)
  --amqpInactiveTimeout  duration      [amqp] When inactive during the given timeout, stop listening ${HTTP_AMQP_INACTIVE_TIMEOUT} (default 0s)
  --amqpMaxRetry         uint          [amqp] Max send retries ${HTTP_AMQP_MAX_RETRY} (default 3)
  --amqpPrefetch         int           [amqp] Prefetch count for QoS ${HTTP_AMQP_PREFETCH} (default 1)
  --amqpQueue            string        [amqp] Queue name ${HTTP_AMQP_QUEUE} (default "httputils")
  --amqpRetryInterval    duration      [amqp] Interval duration when send fails ${HTTP_AMQP_RETRY_INTERVAL} (default 10s)
  --amqpRoutingKey       string        [amqp] RoutingKey name ${HTTP_AMQP_ROUTING_KEY} (default "local")
  --amqpURI              string        [amqp] Address in the form amqps?://<user>:<password>@<address>:<port>/<vhost> ${HTTP_AMQP_URI}
  --cert                 string        [server] Certificate file ${HTTP_CERT}
  --corsCredentials                    [cors] Access-Control-Allow-Credentials ${HTTP_CORS_CREDENTIALS} (default false)
  --corsExpose           string        [cors] Access-Control-Expose-Headers ${HTTP_CORS_EXPOSE}
  --corsHeaders          string        [cors] Access-Control-Allow-Headers ${HTTP_CORS_HEADERS} (default "Content-Type")
  --corsMethods          string        [cors] Access-Control-Allow-Methods ${HTTP_CORS_METHODS} (default "GET")
  --corsOrigin           string        [cors] Access-Control-Allow-Origin ${HTTP_CORS_ORIGIN} (default "*")
  --csp                  string        [owasp] Content-Security-Policy ${HTTP_CSP} (default "default-src 'self'; base-uri 'self'; script-src 'httputils-nonce'")
  --frameOptions         string        [owasp] X-Frame-Options ${HTTP_FRAME_OPTIONS} (default "deny")
  --graceDuration        duration      [http] Grace duration when signal received ${HTTP_GRACE_DURATION} (default 30s)
  --hsts                               [owasp] Indicate Strict Transport Security ${HTTP_HSTS} (default true)
  --idleTimeout          duration      [server] Idle Timeout ${HTTP_IDLE_TIMEOUT} (default 2m0s)
  --key                  string        [server] Key file ${HTTP_KEY}
  --loggerJson                         [logger] Log format as JSON ${HTTP_LOGGER_JSON} (default false)
  --loggerLevel          string        [logger] Logger level ${HTTP_LOGGER_LEVEL} (default "INFO")
  --loggerLevelKey       string        [logger] Key for level in JSON ${HTTP_LOGGER_LEVEL_KEY} (default "level")
  --loggerMessageKey     string        [logger] Key for message in JSON ${HTTP_LOGGER_MESSAGE_KEY} (default "msg")
  --loggerTimeKey        string        [logger] Key for timestamp in JSON ${HTTP_LOGGER_TIME_KEY} (default "time")
  --name                 string        [server] Name ${HTTP_NAME} (default "http")
  --okStatus             int           [http] Healthy HTTP Status code ${HTTP_OK_STATUS} (default 204)
  --port                 uint          [server] Listen port (0 to disable) ${HTTP_PORT} (default 1080)
  --pprofAgent           string        [pprof] URL of the Datadog Trace Agent (e.g. http://datadog.observability:8126) ${HTTP_PPROF_AGENT}
  --pprofPort            int           [pprof] Port of the HTTP server (0 to disable) ${HTTP_PPROF_PORT} (default 0)
  --readTimeout          duration      [server] Read Timeout ${HTTP_READ_TIMEOUT} (default 5s)
  --redisAddress         string slice  [redis] Redis Address host:port (blank to disable) ${HTTP_REDIS_ADDRESS}, as a string slice, environment variable separated by "," (default [127.0.0.1:6379])
  --redisDatabase        int           [redis] Redis Database ${HTTP_REDIS_DATABASE} (default 0)
  --redisMinIdleConn     int           [redis] Redis Minimum Idle Connections ${HTTP_REDIS_MIN_IDLE_CONN} (default 0)
  --redisPassword        string        [redis] Redis Password, if any ${HTTP_REDIS_PASSWORD}
  --redisPoolSize        int           [redis] Redis Pool Size (default GOMAXPROCS*10) ${HTTP_REDIS_POOL_SIZE} (default 0)
  --redisUsername        string        [redis] Redis Username, if any ${HTTP_REDIS_USERNAME}
  --rendererMinify                     [renderer] Minify HTML ${HTTP_RENDERER_MINIFY} (default true)
  --rendererPathPrefix   string        [renderer] Root Path Prefix ${HTTP_RENDERER_PATH_PREFIX}
  --rendererPublicURL    string        [renderer] Public URL ${HTTP_RENDERER_PUBLIC_URL} (default "http://localhost:1080")
  --rendererTitle        string        [renderer] Application title ${HTTP_RENDERER_TITLE} (default "App")
  --shutdownTimeout      duration      [server] Shutdown Timeout ${HTTP_SHUTDOWN_TIMEOUT} (default 10s)
  --telemetryRate        string        [telemetry] OpenTelemetry sample rate, 'always', 'never' or a float value ${HTTP_TELEMETRY_RATE} (default "always")
  --telemetryURL         string        [telemetry] OpenTelemetry gRPC endpoint (e.g. otel-exporter:4317) ${HTTP_TELEMETRY_URL}
  --telemetryUint64                    [telemetry] Change OpenTelemetry Trace ID format to an unsigned int 64 ${HTTP_TELEMETRY_UINT64} (default true)
  --url                  string        [alcotest] URL to check ${HTTP_URL}
  --userAgent            string        [alcotest] User-Agent for check ${HTTP_USER_AGENT} (default "Alcotest")
  --writeTimeout         duration      [server] Write Timeout ${HTTP_WRITE_TIMEOUT} (default 10s)
```
