package main

import (
	"fmt"
	"os"
	"time"

	"github.com/obada-foundation/client-helper/cmd"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/logger"
	"go.uber.org/zap"

	flags "github.com/jessevdk/go-flags"
)

var revision = "unknown"

type opts struct {
	DBPath    string            `long:"db-path" env:"DB_PATH" description:"Show verbose debug information" default:"/home/client-helper/data"`
	ServerCmd cmd.ServerCommand `command:"server"`
}

func main() {
	fmt.Printf("Client Helper %s\n(c) OBADA Foundation %d\n\n", revision, time.Now().Year())

	log, err := logger.New("CLIENT-HELPER")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	var o opts

	if err := run(log, &o); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Errorw("startup", "ERROR", err)
			log.Sync()
			os.Exit(1)
		}
	}
}

func run(logger *zap.SugaredLogger, o *opts) error {
	p := flags.NewParser(o, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		db, err := db.NewDB("client-helper", db.BadgerDBBackend, o.DBPath)
		if err != nil {
			return err
		}
		defer db.Close()

		c := command.(cmd.CommonOptionsCommander)
		c.SetCommon(cmd.CommonOpts{
			Revision: revision,
			Logger:   logger,
			DB:       db,
		})

		err = c.Execute(args)
		if err != nil {
			logger.Errorf("failed with %+v", err)
		}
		return err
	}

	if _, err := p.Parse(); err != nil {
		return err
	}

	return nil
}
