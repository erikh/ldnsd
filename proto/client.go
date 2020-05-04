package proto

import (
	"github.com/erikh/go-transport"
	"github.com/pkg/errors"
)

// NewClient constructs a client from the host and certificate parameters.
func NewClient(host, ca, certName, key string) (DNSControlClient, error) {
	cert, err := transport.LoadCert(ca, certName, key, "")
	if err != nil {
		return nil, errors.Wrap(err, "while loading client certificate")
	}

	cc, err := transport.GRPCDial(cert, host)
	if err != nil {
		return nil, errors.Wrap(err, "while configuring grpc client")
	}

	return NewDNSControlClient(cc), nil
}
