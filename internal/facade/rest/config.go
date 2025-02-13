package rest

type Config struct {
	Port                       uint64 `env:"PORT" yaml:"port"`
	FiberReadTimeout           int64  `env:"FIBER_READ_TIMEOUT" yaml:"fiber-read-timeout"`
	FiberWriteTimeout          int64  `env:"FIBER_WRITE_TIMEOUT" yaml:"fiber-write-timeout"`
	FiberIdleTimeout           int64  `env:"FIBER_IDLE_TIMEOUT" yaml:"fiber-idle-timeout"`
	FiberBodyLimit             int64  `env:"FIBER_BODY_LIMIT" yaml:"fiber-body-limit"`
	FiberReadBufferSize        int64  `env:"FIBER_READ_BUFFER_SIZE" yaml:"fiber-read-buffer-size"`
	FiberStrictRouting         bool   `env:"FIBER_STRICT_ROUTING" yaml:"fiber-strict-routing"`
	FiberCaseSensitive         bool   `env:"FIBER_CASE_SENSITIVE" yaml:"fiber-case-sensitive"`
	FiberDisableStartupMessage bool   `env:"FIBER_DISABLE_STARTUP_MESSAGE" yaml:"fiber-disable-startup-message"`
	FiberDisableKeepalive      bool   `env:"FIBER_DISABLE_KEEPALIVE" yaml:"fiber-disable-keepalive"`
	IsAdditionalErrorsEnabled  bool   `env:"IS_ADDITIONAL_ERRORS_ENABLED" yaml:"is-additional-errors-enabled"`
}
