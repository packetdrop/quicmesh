package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	quicnet "github.com/packetdrop/quicnet/internal"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	qnetLogEnv     = "QNET_LOGLEVEL"
	tunnelOptions      = "Tunnel Options"
	qnetTunnelPort     = 51830
)

func qnetRun(cCtx *cli.Context, logger *zap.Logger) error {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	qnet, err := quicnet.NewQuicNet(
		logger.Sugar(),
		cCtx.String("local-ip"),
		cCtx.String("peer-ip"),
		qnetTunnelPort,
	)
	if err != nil {
		logger.Fatal(err.Error())
	}

	wg := &sync.WaitGroup{}

	if err := qnet.Start(ctx, wg); err != nil {
		logger.Fatal(err.Error())
	}
	<-ctx.Done()
	qnet.Stop()
	wg.Wait()

	return nil
}

// https://www.rfc-editor.org/rfc/rfc9221.html
func main() {
	// set the log level
	debug := os.Getenv(qnetLogEnv)
	var logger *zap.Logger
	var err error
	if debug != "" {
		logger, err = zap.NewDevelopment()
		logger.Info("Debug logging enabled")
	} else {
		logCfg := zap.NewProductionConfig()
		logCfg.DisableStacktrace = true
		logger, err = logCfg.Build()
	}
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Overwrite usage to capitalize "Show"
	cli.HelpFlag.(*cli.BoolFlag).Usage = "Show help"
	// flags are stored in the global flags variable
	app := &cli.App{
		Name:      "qnet",
		Usage:     "Agent to configure encrypted mesh networking using QUIC protocol.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "local-ip",
				Value:    "",
				Usage:    "IP address for the local tunnel interface.",
				Required: true,
				Category: tunnelOptions,
			},
			&cli.StringFlag{
				Name:     "peer-ip",
				Value:    "",
				Usage:    "IP address for the remote peer interface.",
				Required: true,
				Category: tunnelOptions,
			},
		},
		Action: func(cCtx *cli.Context) error {
			return qnetRun(cCtx, logger)
		},

	}
	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}