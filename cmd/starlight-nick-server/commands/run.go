package commands

import (
	"github.com/boreq/guinea"
	"github.com/boreq/starlight-nick-server/config"
	"github.com/boreq/starlight-nick-server/data"
	"github.com/boreq/starlight-nick-server/server"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{
			Name:        "config",
			Optional:    false,
			Multiple:    false,
			Description: "Config file",
		},
	},
	ShortDescription: "runs the server",
}

func runRun(c guinea.Context) error {
	conf, err := config.Load(c.Arguments[0])
	if err != nil {
		return err
	}

	repository, err := data.NewBoltRepository(conf.DatabasePath)
	if err != nil {
		return err
	}

	return server.Serve(repository, conf.ServeAddress)
}
