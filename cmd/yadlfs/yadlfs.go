package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/alxarno/yadlfs/internal"
	"github.com/alxarno/yadlfs/pkg"
	"github.com/urfave/cli/v2"
)

//nolint:gochecknoglobals //variables used in build time for args passing
var (
	Version        = "n/a"
	CommitHash     = "n/a"
	BuildTimestamp = "n/a"
)

func action(_ *cli.Context) error {
	config, err := internal.LoadConfig()
	if err != nil {
		panic(err)
	}

	tmpFolder := ".yadlfs"
	messages := make(chan internal.DialMessage)

	yandexDiskClient := pkg.NewYandexDiskClient(config.YandexDiskOAuthToken, config.YandexDiskProjectFolder)
	dial := internal.NewDial(os.Stdout, messages)
	controller := internal.NewController(yandexDiskClient, tmpFolder, messages)
	dispatcher := internal.NewDispatcher(os.Stdin, controller)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go dial.ListenAndServe(ctx)

	if err := dispatcher.ListenAndServe(ctx); err != nil {
		panic(err)
	}

	return nil
}

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		slog.Info(
			fmt.Sprintf(
				"Version=%s\nCommit-Hash=%s\nBuild-Time=%s\n",
				cCtx.App.Version,
				CommitHash,
				BuildTimestamp,
			),
		)
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "print-version",
		Aliases: []string{"v"},
		Usage:   "print only the version",
	}

	app := &cli.App{
		Name:        "yadlfs",
		Usage:       "git-lfs custom transfer agent which simply works with yandex.disk",
		Version:     Version,
		Copyright:   "(c) github.com/alxarno/yadlfs",
		Suggest:     true,
		HideVersion: false,
		UsageText:   "yadlfs",
		Authors: []*cli.Author{
			{
				Name:  "alxarno",
				Email: "alexarnowork@gmail.com",
			},
		},
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
