package httpsrv

import (
	"errors"
)

type ConfigTLS struct {
	Config
	KeyFile  string
	CertFile string
}

func NewConfigTLS() ConfigTLS {
	cfg := NewConfig()
	cfg.Addr = ":8443"

	return ConfigTLS{
		Config: cfg,
	}
}

func (c *ConfigTLS) Validate() error {
	if c.KeyFile == "" {
		return errors.New("key file is not provided")
	}

	if c.CertFile == "" {
		return errors.New("cert file is not provided")
	}

	return nil
}
