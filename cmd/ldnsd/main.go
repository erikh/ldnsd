package main

import (
	"fmt"
	"net"
	"os"

	"code.hollensbe.org/erikh/ldnsd/dnsdb"
	"code.hollensbe.org/erikh/ldnsd/proto"
	"code.hollensbe.org/erikh/ldnsd/version"
	"github.com/erikh/dnsserver"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	// Author is me
	Author = "Erik Hollensbe <erik+git@hollensbe.org>"
)

func main() {
	app := cli.NewApp()
	app.Version = version.Version
	app.Usage = "Light DNSd -- a small DNS server with a remote control plane"
	app.Author = Author

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "domain, d",
			Usage: "Name of FQDN to put records underneath. Trailing dot will be added automatically.",
			Value: "internal",
		},
	}

	app.Action = runDNS

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func runDNS(ctx *cli.Context) error {
	db, err := dnsdb.New("test.db")
	if err != nil {
		return errors.Wrap(err, "could not open database")
	}

	proto.Boot(db)

	srv := dnsserver.NewWithDB(ctx.GlobalString("domain"), db)
	srv.DeleteA("foo")
	srv.SetA("foo", net.ParseIP("1.2.3.4"))

	return srv.Listen("127.0.0.1:5380")
}
