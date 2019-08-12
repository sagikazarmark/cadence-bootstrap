package cadence

import (
	"fmt"

	"emperror.dev/errors"
)

// Config holds information necessary for connecting to Cadence.
type Config struct {
	// Cadence connection details
	Host string
	Port int

	Domains []Domain
}

// Domain is a Cadence domain.
type Domain struct {
	Name                                   string
	Description                            string
	WorkflowExecutionRetentionPeriodInDays int32
	EmitMetric                             bool
}

func (d Domain) Validate() error {
	if d.Name == "" {
		return errors.New("cadence domain name is mandatory")
	}

	if d.WorkflowExecutionRetentionPeriodInDays < 1 {
		return errors.New("cadence domain retention period is mandatory")
	}

	return nil
}

func (c Config) Peer() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("cadence host is required")
	}

	if c.Port == 0 {
		return errors.New("cadence port is required")
	}

	for _, domain := range c.Domains {
		err := domain.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}
