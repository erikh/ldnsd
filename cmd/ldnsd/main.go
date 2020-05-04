package main

import (
	"fmt"
	"os"

	"code.hollensbe.org/erikh/ldnsd/config"
	"code.hollensbe.org/erikh/ldnsd/service"
	"code.hollensbe.org/erikh/ldnsd/version"
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
	app.UsageText = app.Name + " [options] [config file]"
	app.Author = Author
	app.Action = runDNS

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func runDNS(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("invalid arguments")
	}

	c, err := config.Parse(ctx.Args()[0])
	if err != nil {
		return errors.Wrap(err, "while parsing configuration")
	}

	if err := service.Boot(ctx.App.Name, c); err != nil {
		return errors.Wrap(err, "while running service")
	}

	return nil
}
