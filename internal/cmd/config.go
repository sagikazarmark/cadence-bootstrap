package main

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/sagikazarmark/cadence-bootstrap/internal/platform/backoffpolicy"
	"github.com/sagikazarmark/cadence-bootstrap/internal/platform/cadence"
	"github.com/sagikazarmark/cadence-bootstrap/internal/platform/log"
)

// configuration holds any kind of configuration that comes from the outside world and
// is necessary for running the application.
type configuration struct {
	// Log configuration
	Log log.Config

	Backoff backoffpolicy.Config

	Cadence cadence.Config
}

// Validate validates the configuration.
func (c configuration) Validate() error {
	if err := c.Backoff.Validate(); err != nil {
		return err
	}

	if err := c.Cadence.Validate(); err != nil {
		return err
	}

	return nil
}

// configure configures some defaults in the Viper instance.
func configure(v *viper.Viper, _ *pflag.FlagSet) {
	// Viper settings
	v.AddConfigPath(".")

	// Environment variable settings
	// v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	// Application constants
	v.Set("appName", appName)

	// Global configuration
	v.SetDefault("debug", false)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		v.SetDefault("no_color", true)
	}

	// Log configuration
	v.SetDefault("log.format", "json")
	v.SetDefault("log.level", "info")
	v.RegisterAlias("log.noColor", "no_color")

	// Backoff configuration
	v.SetDefault("backoff.policy", "exponential")
	v.SetDefault("backoff.maxRetries", 10)
	v.SetDefault("backoff.delay", 1*time.Second)
	v.SetDefault("backoff.interval", 500*time.Millisecond)
	v.SetDefault("backoff.maxInterval", 2*time.Minute)
	v.SetDefault("backoff.jitterFactor", 0.5)
	v.SetDefault("backoff.factor", float64(2))

	// Cadence config
	_ = v.BindEnv("cadence.host")
	v.SetDefault("cadence.port", 7933)
	v.SetDefault("cadence.domains", nil)
}
