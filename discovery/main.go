package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
)

var (
	logger = logrus.New()
)

func getEtcdClient(context *cli.Context) *etcd.Client {
	return etcd.NewClient(context.GlobalStringSlice("etcd"))
}

func main() {
	app := cli.NewApp()
	app.Name = "discovery"
	app.Usage = "automatic discovery for docker clusters"
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{Name: "etcd", Value: &cli.StringSlice{}, Usage: "list of etcd machines"},
		cli.BoolFlag{Name: "debug", Usage: "enable debug output"},
	}

	app.Commands = []cli.Command{
		serveCommand,
	}

	app.Before = func(context *cli.Context) error {
		if context.GlobalBool("debug") {
			logger.Level = logrus.DebugLevel
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
