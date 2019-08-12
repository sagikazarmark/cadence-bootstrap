package backoffpolicy

import (
	"time"

	"emperror.dev/errors"
	"github.com/lestrrat-go/backoff"
)

const (
	constantPolicy    = "constant"
	exponentialPolicy = "exponential"
)

// Config contains
type Config struct {
	Policy string

	// Common
	MaxRetries     int
	MaxElapsedTime time.Duration

	// Constant
	Delay time.Duration

	// Exponential
	Interval     time.Duration
	MaxInterval  time.Duration
	JitterFactor float64
	Factor       float64
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.Policy != constantPolicy && c.Policy != exponentialPolicy {
		return errors.New("backoff policy must be 'constant' or 'exponential'")
	}

	if c.Policy == constantPolicy {
		return errors.New("constant backoff policy requires a constant delay")
	}

	return nil
}

// NewPolicy creates a new backoff policy from config.
func NewPolicy(config Config) backoff.Policy {
	options := []backoff.Option{
		backoff.WithMaxRetries(config.MaxRetries),
		backoff.WithMaxElapsedTime(config.MaxElapsedTime),
		backoff.WithInterval(config.Interval),
		backoff.WithMaxInterval(config.MaxInterval),
		backoff.WithJitterFactor(config.JitterFactor),
		backoff.WithFactor(config.Factor),
	}

	switch config.Policy {
	case constantPolicy:
		return backoff.NewConstant(config.Delay, options...)

	case exponentialPolicy:
		return backoff.NewExponential(options...)

	default:
		panic("unknown policy: " + config.Policy)
	}
}
