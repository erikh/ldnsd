package config

import (
	"io/ioutil"

	"github.com/erikh/go-transport"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	defaultDBFile   = "ldnsd.db"
	defaultCAFile   = "/etc/ldnsd/rootCA.pem"
	defaultCertFile = "/etc/ldnsd/server.pem"
	defaultKeyFile  = "/etc/ldnsd/server.key"
)

// Config is the configuration of the dhcpd service
type Config struct {
	DBFile      string      `yaml:"db_file"`
	Certificate Certificate `yaml:"certificate"`
}

// Parse parses the configuration in the file and returns it.
func Parse(filename string) (Config, error) {
	var config Config

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, errors.Wrap(err, "while reading configuration file")
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return config, errors.Wrap(err, "error while parsing configuration file")
	}

	return config, config.validateAndFix()
}

func (c *Config) validateAndFix() error {
	if c.Certificate.CertFile == "" {
		c.Certificate.CertFile = defaultCertFile
	}

	if c.Certificate.KeyFile == "" {
		c.Certificate.KeyFile = defaultKeyFile
	}

	if c.Certificate.CAFile == "" {
		c.Certificate.CAFile = defaultCAFile
	}

	return nil
}

// Certificate iconifies the certificate used to authenticate GRPC connections.
type Certificate struct {
	CAFile   string `yaml:"ca"`
	CertFile string `yaml:"cert"`
	KeyFile  string `yaml:"key"`
}

// NewCert returns a transport interface suitable for use with GRPC servers.
func (crt Certificate) NewCert() (*transport.Cert, error) {
	return transport.LoadCert(crt.CAFile, crt.CertFile, crt.KeyFile, "")
}
