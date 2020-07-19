package service

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/erikh/dnsserver"
	"github.com/erikh/go-transport"
	"github.com/erikh/ldnsd/config"
	"github.com/erikh/ldnsd/dnsdb"
	"github.com/erikh/ldnsd/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Service is the encapsulation of a fully composed service.
type Service struct {
	config  *config.Config
	appName string
	grpcS   *grpc.Server
	l       net.Listener
	handler *dnsserver.Server
}

// InstallSignalHandler installs a signal handler that allows it to trap exit
// conditions to react to them, like gracefully shutting down.
func (s *Service) InstallSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	go func() {
		for {
			switch <-sigChan {
			// FIXME add config reload as SIGUSR1 or SIGHUP
			case syscall.SIGTERM, syscall.SIGINT:
				s.Shutdown()
				os.Exit(0)
			}
		}
	}()
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
}

// New constructs a new service from a config.Config
func New(name string, c *config.Config) (*Service, error) {
	db, err := dnsdb.New(c.DBFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not open database")
	}

	cert, err := c.Certificate.NewCert()
	if err != nil {
		return nil, errors.Wrap(err, "invalid certificate configuration")
	}

	srv := dnsserver.NewWithDB(c.Domain, db)
	grpcS := proto.Boot(srv)
	l, err := transport.Listen(cert, "tcp", c.GRPCListen)
	if err != nil {
		return nil, errors.Wrap(err, "while configuring grpc listener")
	}

	return &Service{
		l:       l,
		grpcS:   grpcS,
		handler: srv,
		appName: name,
		config:  c,
	}, nil
}

// Shutdown the service.
func (s *Service) Shutdown() {
	logrus.Infof("Stopping %v...", s.appName)
	s.grpcS.GracefulStop()
	s.l.Close()
	s.handler.Close()
	logrus.Infof("Done.")
}

// Boot the service
func (s *Service) Boot() error {
	go s.grpcS.Serve(s.l)
	return s.handler.Listen(s.config.DNSListen)
}
