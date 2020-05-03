package main

import (
	"context"
	"fmt"
	"os"

	"code.hollensbe.org/erikh/ldnsd/proto"
	"code.hollensbe.org/erikh/ldnsd/version"
	"github.com/erikh/go-transport"
	"github.com/golang/protobuf/ptypes/empty"
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
	app.Author = Author

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, t",
			Usage: "Set the host:port connection for GRPC",
			Value: "localhost:7847",
		},
		cli.StringFlag{
			Name:  "cert, c",
			Usage: "Set the client certificate for authentication",
			Value: "/etc/ldnsd/client.pem",
		},
		cli.StringFlag{
			Name:  "key, k",
			Usage: "Set the client certificate key",
			Value: "/etc/ldnsd/client.key",
		},
		cli.StringFlag{
			Name:  "ca",
			Usage: "Set the certificate authority",
			Value: "/etc/ldnsd/rootCA.pem",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "list",
			Action: list,
			Usage:  "List the A record table",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func getClient(ctx *cli.Context) (proto.DNSControlClient, error) {
	cert, err := transport.LoadCert(ctx.GlobalString("ca"), ctx.GlobalString("cert"), ctx.GlobalString("key"), "")
	if err != nil {
		return nil, errors.Wrap(err, "while loading client certificate")
	}

	cc, err := transport.GRPCDial(cert, ctx.GlobalString("host"))
	if err != nil {
		return nil, errors.Wrap(err, "while configuring grpc client")
	}

	return proto.NewDNSControlClient(cc), nil
}

func list(ctx *cli.Context) error {
	client, err := getClient(ctx)
	if err != nil {
		return errors.Wrap(err, "could not create client")
	}

	list, err := client.ListA(context.Background(), &empty.Empty{})
	if err != nil {
		return errors.Wrap(err, "cold not query A record list")
	}

	fmt.Println("Host\tIP")

	for _, record := range list.Records {
		fmt.Printf("%s\t%s\n", record.Host, record.Address)
	}

	return nil
}
