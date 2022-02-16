package main

import (
	"github.com/bryant-rh/kubectl-resource/cmd"
	"github.com/bryant-rh/kubectl-resource/pkg/util"

	"github.com/spf13/pflag"
)

func init() {
	flags := pflag.NewFlagSet("kubectl-resource", pflag.ExitOnError)
	pflag.CommandLine = flags
}

func main() {
	command := cmd.NewCmdResource()
	util.CheckErr(command.Execute())
}
