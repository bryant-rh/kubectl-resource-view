package main

import (
	cmd "github.com/bryant-rh/kubectl-resource-view/cmd/kubectl-resource-view"
	"github.com/bryant-rh/kubectl-resource-view/pkg/util"

	"github.com/spf13/pflag"
)

func init() {
	flags := pflag.NewFlagSet("kubectl-resource-view", pflag.ExitOnError)
	pflag.CommandLine = flags
}

func main() {
	command := cmd.NewCmdResource()
	util.CheckErr(command.Execute())
}
