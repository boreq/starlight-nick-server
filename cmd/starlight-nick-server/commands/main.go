package commands

import (
	"github.com/boreq/guinea"
)

var MainCmd = guinea.Command{
	Run: runMain,
	Subcommands: map[string]*guinea.Command{
		"run":            &runCmd,
		"default_config": &defaultConfigCmd,
	},
	ShortDescription: "a nick server for starlight",
	Description: `
This server provides human-readable nicknames for starlight. Each node can
insert a mapping linking a human-readable nick with that node's id into the
system and other nodes can recover it later.
`,
}

func runMain(c guinea.Context) error {
	return guinea.ErrInvalidParms
}
