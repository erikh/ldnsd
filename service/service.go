package service

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"code.hollensbe.org/erikh/ldnsd/config"
	"code.hollensbe.org/erikh/ldnsd/dnsdb"
	"code.hollensbe.org/erikh/ldnsd/proto"
	"github.com/erikh/dnsserver"
	"github.com/erikh/go-transport"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func installSignalHandler(appName string, grpcS *grpc.Server, l net.Listener, handler *dnsserver.Server) {
	sigChan := make(chan os.Signal, 1)
	go func() {
		for {
			switch <-sigChan {
			// FIXME add config reload as SIGUSR1 or SIGHUP
			case syscall.SIGTERM, syscall.SIGINT:
				logrus.Infof("Stopping %v...", appName)
				grpcS.GracefulStop()
				l.Close()
				handler.Close()
				logrus.Infof("Done.")
				os.Exit(0)
			}
		}
	}()
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
}

// Boot the service
func Boot(name string, c config.Config) error {
	db, err := dnsdb.New(c.DBFile)
	if err != nil {
		return errors.Wrap(err, "could not open database")
	}

	cert, err := c.Certificate.NewCert()
	if err != nil {
		return errors.Wrap(err, "invalid certificate configuration")
	}

	srv := dnsserver.NewWithDB(c.Domain, db)
	grpcS := proto.Boot(srv)
	l, err := transport.Listen(cert, "tcp", c.GRPCListen)
	if err != nil {
		return errors.Wrap(err, "while configuring grpc listener")
	}

	go grpcS.Serve(l)
	installSignalHandler(name, grpcS, l, srv)

	return srv.Listen(c.DNSListen)
}
