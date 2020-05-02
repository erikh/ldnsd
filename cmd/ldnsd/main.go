package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"code.hollensbe.org/erikh/ldnsd/config"
	"code.hollensbe.org/erikh/ldnsd/dnsdb"
	"code.hollensbe.org/erikh/ldnsd/proto"
	"code.hollensbe.org/erikh/ldnsd/version"
	"github.com/erikh/dnsserver"
	"github.com/erikh/go-transport"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

const (
	// Author is me
	Author = "Erik Hollensbe <erik+git@hollensbe.org>"
)

func main() {
	app := cli.NewApp()
	app.Version = version.Version
	app.Usage = "Light DNSd -- a small DNS server with a remote control plane"
	app.ArgsUsage = "[options] [config file]"
	app.Author = Author

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "domain, d",
			Usage: "Name of FQDN to put records underneath. Trailing dot will be added automatically.",
			Value: "internal",
		},
		cli.StringFlag{
			Name:  "listen, l",
			Usage: "Change the host:port to listen for GRPC connections",
			Value: "localhost:7847",
		},
	}

	app.Action = runDNS

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

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

func runDNS(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("invalid arguments")
	}

	c, err := config.Parse(ctx.Args()[0])
	if err != nil {
		return errors.Wrap(err, "while parsing configuration")
	}

	db, err := dnsdb.New(c.DBFile)
	if err != nil {
		return errors.Wrap(err, "could not open database")
	}

	srv := dnsserver.NewWithDB(ctx.GlobalString("domain"), db)
	srv.DeleteA("foo")
	srv.SetA("foo", net.ParseIP("1.2.3.4"))

	cert, err := c.Certificate.NewCert()
	if err != nil {
		return errors.Wrap(err, "invalid certificate configuration")
	}

	grpcS := proto.Boot(db)
	l, err := transport.Listen(cert, "tcp", ctx.GlobalString("listen"))
	if err != nil {
		return errors.Wrap(err, "while configuring grpc listener")
	}
	go grpcS.Serve(l)

	installSignalHandler(ctx.App.Name, grpcS, l, srv)

	return srv.Listen("127.0.0.1:5380")
}
