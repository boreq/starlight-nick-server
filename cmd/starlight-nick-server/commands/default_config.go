package commands

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/guinea"
	"github.com/boreq/starlight-nick-server/config"
)

var defaultConfigCmd = guinea.Command{
	Run:              runDefaultConfig,
	ShortDescription: "prints the default config",
}

func runDefaultConfig(c guinea.Context) error {
	conf := config.Default()

	j, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", j)
	return nil
}
