package presidioprocessor

import (
	"errors"

	"go.opentelemetry.io/collector/config/configgrpc"
)

// Config represents the processor config settings within config.yaml
type Config struct {
	// Embeds standard gRPC client settings (Timeout, Headers, TLS, etc.)
	configgrpc.ClientConfig `mapstructure:",squash"`

	// Attributes is a list of trace/log attribute keys to inspect and mask
	Attributes []string `mapstructure:"attributes"`

	IncludeLogBody bool `mapstructure:"include_log_body"`
}

// Validate checks if the receiver configuration is valid
func (c *Config) Validate() error {

	if len(c.Attributes) == 0 {
		return errors.New("at least one target attribute must be specified")
	}

	return nil
}
