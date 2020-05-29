package main

import (
	"os"

	"mydocker/command"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = ` mydocker usage.`

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage
	app.Commands = command.Commands
	app.Before = func(ctx *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
