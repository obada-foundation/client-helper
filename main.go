package main

import (
	"expvar"
	"fmt"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	flags "github.com/jessevdk/go-flags"
	"github.com/obada-foundation/client-helper/cmd"
	"github.com/obada-foundation/client-helper/system/logger"
	"github.com/tendermint/tm-db"
	"go.uber.org/zap"
)

var revision = "unknown"

type opts struct {
	DBPath    string            `long:"db-path" env:"DB_PATH" description:"Show verbose debug information" default:"/home/client-helper/data"`
	ServerCmd cmd.ServerCommand `command:"server"`
}

//nolint:gochecknoinits // this is an entrypoint
func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()
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
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type != flags.ErrHelp {
			log.Errorw("startup", "ERROR", err)
			_ = log.Sync()

			// nolint:gocritic //for future refactoring
			os.Exit(1)
		}
	}
}

func run(lgr *zap.SugaredLogger, o *opts) error {
	expvar.NewString("build").Set(revision)

	p := flags.NewParser(o, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		db, err := db.NewDB("client-helper", db.BadgerDBBackend, o.DBPath)
		if err != nil {
			return err
		}

		c := command.(cmd.CommonOptionsCommander)
		c.SetCommon(cmd.CommonOpts{
			Revision: revision,
			Logger:   lgr,
			DB:       db,
		})

		err = c.Execute(args)
		if err != nil {
			lgr.Errorf("failed with %+v", err)
		}
		return err
	}

	if _, err := p.Parse(); err != nil {
		return err
	}

	return nil
}
