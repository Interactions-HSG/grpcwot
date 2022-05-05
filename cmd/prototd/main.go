package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/Interactions-HSG/grpcwot"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   50051,
				Usage:   "The port for the gRPC service",
			},
			&cli.StringFlag{
				Name:  "ip",
				Value: "127.0.0.1",
				Usage: "The IP address for the gRPC service",
			},
			&cli.StringFlag{
				Name:    "outputDir",
				Aliases: []string{"o"},
				Value:   "output",
				Usage:   "Write the resulting Thing Description and files to `DIR`",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "",
				Usage:   "Load a configuration for affordance classification",
			},
			&cli.StringFlag{
				Name:    "reduced",
				Aliases: []string{"r"},
				Value:   "",
				Usage:   "Load a configuration file to only include specific RPC functions for the TD",
			},
		},
		Name:  "prototd",
		Usage: "Translate ProtocolBuffers to ThingDescription",
		Action: func(c *cli.Context) error {
			protoFile := c.Args().Get(0)
			if !strings.HasSuffix(protoFile, ".proto") {
				return errors.New("the input file must be a .proto file")
			} else if _, err := os.Stat(protoFile); errors.Is(err, os.ErrNotExist) {
				return err
			}
			return grpcwot.GenerateTDfromProtoBuf(
				protoFile,
				c.String("outputDir"),
				c.String("config"),
				c.String("reduced"),
				c.String("ip"),
				c.Int("port"))
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
