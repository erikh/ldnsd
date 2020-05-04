package main

import (
	"context"
	"fmt"
	"os"

	"code.hollensbe.org/erikh/ldnsd/proto"
	"code.hollensbe.org/erikh/ldnsd/version"
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
			Name:      "list",
			ArgsUsage: " ",
			Action:    list,
			Usage:     "List the A record table",
		},
		{
			Name:      "set",
			Action:    set,
			ArgsUsage: "[host] [v4 IP]",
			Usage:     "Set an A record, only takes IPv4",
		},
		{
			Name:      "delete",
			Action:    delete,
			ArgsUsage: "[host]",
			Usage:     "Delete an A record by hostname",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func getClient(ctx *cli.Context) (proto.DNSControlClient, error) {
	return proto.NewClient(
		ctx.GlobalString("host"),
		ctx.GlobalString("ca"),
		ctx.GlobalString("cert"),
		ctx.GlobalString("key"),
	)
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

func set(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return errors.New("invalid arguments")
	}

	client, err := getClient(ctx)
	if err != nil {
		return errors.Wrap(err, "could not create client")
	}

	_, err = client.SetA(context.Background(), &proto.Record{
		Host:    ctx.Args()[0],
		Address: ctx.Args()[1],
	})

	if err != nil {
		return errors.Wrap(err, "could not set A record")
	}

	return nil
}

func delete(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("invalid arguments")
	}

	client, err := getClient(ctx)
	if err != nil {
		return errors.Wrap(err, "could not create client")
	}

	_, err = client.DeleteA(context.Background(), &proto.Record{Host: ctx.Args()[0]})
	if err != nil {
		return errors.Wrap(err, "could not set A record")
	}

	return nil
}
