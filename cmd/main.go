package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/inaccel/reef/internal"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var version string

func main() {
	app := &cli.App{
		Name:    "reef",
		Version: version,
		Usage:   "A self-sufficient runtime for accelerators.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "cert",
				Usage: "SSL certification file",
				Value: "/etc/inaccel/certs/ssl.pem",
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "SSL key file",
				Value: "/etc/inaccel/private/ssl.key",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug output",
			},
		},
		Before: func(context *cli.Context) error {
			log.SetOutput(io.Discard)

			logrus.SetFormatter(new(logrus.JSONFormatter))

			if context.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			return http.ListenAndServeTLS("", context.String("cert"), context.String("key"), internal.Handle("/"))
		},
		Commands: []*cli.Command{
			initCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
